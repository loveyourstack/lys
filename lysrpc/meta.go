package lysrpc

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/lystype"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var timeType = reflect.TypeFor[time.Time]()
var uuidType = reflect.TypeFor[uuid.UUID]()

// lystypeLayouts lists the lystype time types, which are carried as proto strings in their json layout
var lystypeLayouts = map[reflect.Type]string{
	reflect.TypeFor[lystype.Date]():     lystype.DateFormat,
	reflect.TypeFor[lystype.Time]():     lystype.TimeFormat,
	reflect.TypeFor[lystype.Datetime](): lystype.DatetimeFormat,
}

const timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"

// pbFieldMap links a proto field to the Go struct field (by index path) with the same json name
type pbFieldMap struct {
	fd  protoreflect.FieldDescriptor
	idx []int
}

// jsonFieldIndexes returns the index path of each exported struct field by json name. Embedded structs are flattened
func jsonFieldIndexes(t reflect.Type) map[string][]int {
	res := make(map[string][]int)
	for _, f := range reflect.VisibleFields(t) {
		if f.Anonymous || !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		res[name] = f.Index
	}
	return res
}

func isTimestamp(fd protoreflect.FieldDescriptor) bool {
	return fd.Message() != nil && fd.Message().FullName() == timestampFullName
}

// buildPbFieldMaps matches every proto field to a Go field of goT. Panics on any mismatch, so drift is found at startup
func buildPbFieldMaps(pbDesc protoreflect.MessageDescriptor, goT reflect.Type, requireAllGoFields bool) []pbFieldMap {
	goIdx := jsonFieldIndexes(goT)
	fields := pbDesc.Fields()
	maps := make([]pbFieldMap, 0, fields.Len())
	used := make(map[string]bool, fields.Len())

	for i := range fields.Len() {
		fd := fields.Get(i)
		name := string(fd.Name())
		idx, ok := goIdx[name]
		if !ok {
			panic(fmt.Sprintf("proto field %s.%s has no json field in %s", pbDesc.Name(), name, goT))
		}
		if fd.IsMap() {
			panic(fmt.Sprintf("proto field %s.%s: maps are not supported", pbDesc.Name(), name))
		}
		ft := goT.FieldByIndex(idx).Type
		if err := checkFieldCompat(fd, ft); err != nil {
			panic(fmt.Sprintf("proto field %s.%s vs %s: %s", pbDesc.Name(), name, goT, err))
		}
		used[name] = true
		maps = append(maps, pbFieldMap{fd: fd, idx: idx})
	}

	if requireAllGoFields {
		for name := range goIdx {
			if !used[name] {
				panic(fmt.Sprintf("%s field %q has no proto field in %s", goT, name, pbDesc.Name()))
			}
		}
	}
	return maps
}

func checkFieldCompat(fd protoreflect.FieldDescriptor, ft reflect.Type) error {
	if fd.IsList() {
		if ft.Kind() != reflect.Slice {
			return fmt.Errorf("repeated field requires a slice, got %s", ft)
		}
		ft = ft.Elem()
	}
	// nullable model fields are pointers
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	k := ft.Kind()
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		if k >= reflect.Int && k <= reflect.Int64 {
			return nil
		}
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		if k >= reflect.Uint && k <= reflect.Uint64 {
			return nil
		}
	case protoreflect.BoolKind:
		if k == reflect.Bool {
			return nil
		}
	case protoreflect.StringKind, protoreflect.EnumKind:
		if k == reflect.String {
			return nil
		}
		if fd.Kind() == protoreflect.StringKind {
			if _, ok := lystypeLayouts[ft]; ok || ft == uuidType {
				return nil
			}
		}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		if k == reflect.Float32 || k == reflect.Float64 {
			return nil
		}
	case protoreflect.MessageKind:
		if isTimestamp(fd) && ft.ConvertibleTo(timeType) {
			return nil
		}
	}
	return fmt.Errorf("unsupported type mapping %s -> %s", fd.Kind(), ft)
}

// goToPbField sets the proto field from the Go value, handling both scalar and repeated fields
func goToPbField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, v reflect.Value) {
	if !fd.IsList() {
		if pv, ok := goToPbValue(fd, v); ok {
			msg.Set(fd, pv)
		}
		return
	}
	if v.Len() == 0 {
		return
	}
	list := msg.Mutable(fd).List()
	for i := range v.Len() {
		if pv, ok := goToPbValue(fd, v.Index(i)); ok {
			list.Append(pv)
		}
	}
}

// pbToGoField sets the Go value from the proto field, handling both scalar and repeated fields
func pbToGoField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, dst reflect.Value) {
	if !fd.IsList() {
		if msg.Has(fd) {
			pbToGoValue(fd, msg.Get(fd), dst)
		}
		return
	}

	list := msg.Get(fd).List()

	// set non-nil even when empty, so it's written as '{}' not NULL
	// note: this will make NULLable slice fields non-nil even when empty
	sl := reflect.MakeSlice(dst.Type(), list.Len(), list.Len())

	for i := range list.Len() {
		pbToGoValue(fd, list.Get(i), sl.Index(i))
	}
	dst.Set(sl)
}

func goToPbValue(fd protoreflect.FieldDescriptor, v reflect.Value) (protoreflect.Value, bool) {
	// nil pointers are left unset
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return protoreflect.Value{}, false
		}
		v = v.Elem()
	}
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(int32(v.Int())), true
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(v.Int()), true
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(uint32(v.Uint())), true
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(v.Uint()), true
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(v.Bool()), true
	case protoreflect.StringKind:
		if layout, ok := lystypeLayouts[v.Type()]; ok {
			// zero times are left unset
			t := v.Convert(timeType).Interface().(time.Time)
			if t.IsZero() {
				return protoreflect.Value{}, false
			}
			return protoreflect.ValueOfString(t.Format(layout)), true
		}
		if v.Type() == uuidType {
			// zero uuids are left unset
			u := v.Interface().(uuid.UUID)
			if u == uuid.Nil {
				return protoreflect.Value{}, false
			}
			return protoreflect.ValueOfString(u.String()), true
		}
		return protoreflect.ValueOfString(v.String()), true
	case protoreflect.EnumKind:
		// enum value names are used as-is; an empty or unknown name is left unset
		ev := fd.Enum().Values().ByName(protoreflect.Name(v.String()))
		if ev == nil {
			return protoreflect.Value{}, false
		}
		return protoreflect.ValueOfEnum(ev.Number()), true
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(float32(v.Float())), true
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(v.Float()), true
	default: // timestamp: zero times are left unset
		t := v.Convert(timeType).Interface().(time.Time)
		if t.IsZero() {
			return protoreflect.Value{}, false
		}
		return protoreflect.ValueOfMessage(timestamppb.New(t).ProtoReflect()), true
	}
}

func pbToGoValue(fd protoreflect.FieldDescriptor, pv protoreflect.Value, dst reflect.Value) {
	if dst.Kind() == reflect.Pointer {
		p := reflect.New(dst.Type().Elem())
		pbToGoValue(fd, pv, p.Elem())
		dst.Set(p)
		return
	}
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		dst.SetInt(pv.Int())
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		dst.SetUint(pv.Uint())
	case protoreflect.BoolKind:
		dst.SetBool(pv.Bool())
	case protoreflect.StringKind:
		if layout, ok := lystypeLayouts[dst.Type()]; ok {
			// an unparseable time is left as the zero value
			if t, err := time.Parse(layout, pv.String()); err == nil {
				dst.Set(reflect.ValueOf(t).Convert(dst.Type()))
			}
			return
		}
		if dst.Type() == uuidType {
			// an unparseable uuid is left as the zero value
			if u, err := uuid.Parse(pv.String()); err == nil {
				dst.Set(reflect.ValueOf(u))
			}
			return
		}
		dst.SetString(pv.String())
	case protoreflect.EnumKind:
		if ev := fd.Enum().Values().ByNumber(pv.Enum()); ev != nil {
			dst.SetString(string(ev.Name()))
		}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		dst.SetFloat(pv.Float())
	default: // timestamp
		ts := pv.Message().Interface().(*timestamppb.Timestamp)
		dst.Set(reflect.ValueOf(ts.AsTime()).Convert(dst.Type()))
	}
}

// NewToPb returns a func converting a store model to its proto message. Fields are matched by proto field name = model json name.
// Embedded structs are flattened. Repeated fields map to slices. Zero timestamps are left nil. Model fields without a proto field are ignored.
// Panics if a proto field has no matching model field or an unsupported type, so call it at init.
func NewToPb[modelT any, pbT any, pbPtrT interface {
	*pbT
	proto.Message
}]() func(modelT) *pbT {

	desc := pbPtrT(new(pbT)).ProtoReflect().Descriptor()
	maps := buildPbFieldMaps(desc, reflect.TypeFor[modelT](), false)

	return func(m modelT) *pbT {
		out := pbPtrT(new(pbT))
		msg := out.ProtoReflect()
		mv := reflect.ValueOf(m)
		for _, fm := range maps {
			goToPbField(msg, fm.fd, mv.FieldByIndex(fm.idx))
		}
		return out
	}
}

// NewFromPb returns a func converting a proto input message to a store input struct. Fields are matched by proto field name = input json name.
// Panics if the proto fields and the input fields are not an exact match or have unsupported types, so call it at init.
func NewFromPb[inputT any, pbT any, pbPtrT interface {
	*pbT
	proto.Message
}]() func(pbPtrT) inputT {

	desc := pbPtrT(new(pbT)).ProtoReflect().Descriptor()
	maps := buildPbFieldMaps(desc, reflect.TypeFor[inputT](), true)

	return func(p pbPtrT) inputT {
		var in inputT
		iv := reflect.ValueOf(&in).Elem()
		msg := p.ProtoReflect()
		for _, fm := range maps {
			pbToGoField(msg, fm.fd, iv.FieldByIndex(fm.idx))
		}
		return in
	}
}
