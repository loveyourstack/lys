package lysrpc

import (
	"fmt"
	"net/netip"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/lystype"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var netipAddrType = reflect.TypeFor[netip.Addr]()
var timeType = reflect.TypeFor[time.Time]()
var uuidType = reflect.TypeFor[uuid.UUID]()

// lystypeLayouts lists the lystype time types, which are carried as proto strings in their json layout
var lystypeLayouts = map[reflect.Type]string{
	reflect.TypeFor[lystype.Date](): lystype.DateFormat,
	reflect.TypeFor[lystype.Time](): lystype.TimeFormat,
}

const timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"

// pbFieldMap links a proto field to the Go struct field (by index path) with the same json name
type pbFieldMap struct {
	fd  protoreflect.FieldDescriptor
	idx []int
}

// fieldByIndex retrieves the nested struct field specified by the index path.
// If allocate is true, it will allocate any nil pointers along the path.
func fieldByIndex(v reflect.Value, index []int, allocate bool) (reflect.Value, bool) {
	for _, i := range index {
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				if !allocate || !v.CanSet() {
					return reflect.Value{}, false
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		v = v.Field(i)
	}
	return v, true
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

func isSignedInteger(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func isUnsignedInteger(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uint64
}

// checkFieldCompat returns an error if the field's proto type is incompatible with its Go type.
func checkFieldCompat(fd protoreflect.FieldDescriptor, ft reflect.Type) error {

	// repeated fields must be slices in Go
	if fd.IsList() {
		if ft.Kind() != reflect.Slice {
			return fmt.Errorf("repeated field requires a slice, got %s", ft)
		}
		ft = ft.Elem()

		// repeated field elements must not be pointers
		// protobuf repeated fields cannot represent null elements, so the list would be compacted
		if ft.Kind() == reflect.Pointer {
			return fmt.Errorf("repeated field elements cannot be pointers, got %s", ft)
		}
	}

	// nullable model fields are pointers
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}

	k := ft.Kind()
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if isSignedInteger(k) && ft.Bits() == 32 {
			return nil
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if isSignedInteger(k) && ft.Bits() == 64 {
			return nil
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if isUnsignedInteger(k) && ft.Bits() == 32 {
			return nil
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if isUnsignedInteger(k) && ft.Bits() == 64 {
			return nil
		}
	case protoreflect.FloatKind:
		if k == reflect.Float32 {
			return nil
		}
	case protoreflect.DoubleKind:
		if k == reflect.Float64 {
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
			if _, ok := lystypeLayouts[ft]; ok || ft == netipAddrType || ft == uuidType {
				return nil
			}
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
func pbToGoField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, dst reflect.Value) error {
	if !fd.IsList() {
		if msg.Has(fd) {
			if err := pbToGoValue(fd, msg.Get(fd), dst); err != nil {
				return fmt.Errorf("field %s: %w", fd.Name(), err)
			}
		}
		return nil
	}

	list := msg.Get(fd).List()

	// set non-nil even when empty, so it's written as '{}' not NULL
	// note: this will make NULLable slice fields non-nil even when empty
	sl := reflect.MakeSlice(dst.Type(), list.Len(), list.Len())

	for i := range list.Len() {
		if err := pbToGoValue(fd, list.Get(i), sl.Index(i)); err != nil {
			return fmt.Errorf("field %s element %d: %w", fd.Name(), i, err)
		}
	}
	dst.Set(sl)
	return nil
}

// goToPbValue converts a Go value to a proto value, returning the proto value and a boolean indicating if it should be set.
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

		if v.Type() == netipAddrType {
			// zero netip.Addr are left unset
			a := v.Interface().(netip.Addr)
			if !a.IsValid() {
				return protoreflect.Value{}, false
			}
			return protoreflect.ValueOfString(a.String()), true
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

// pbToGoValue converts a proto value to a Go value, setting the destination reflect.Value accordingly.
func pbToGoValue(fd protoreflect.FieldDescriptor, pv protoreflect.Value, dst reflect.Value) error {

	if dst.Kind() == reflect.Pointer {
		p := reflect.New(dst.Type().Elem())
		if err := pbToGoValue(fd, pv, p.Elem()); err != nil {
			return err
		}
		dst.Set(p)
		return nil
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
			t, err := time.Parse(layout, pv.String())
			if err != nil {
				return fmt.Errorf("invalid %s value %q: %w", dst.Type(), pv.String(), err)
			}
			dst.Set(reflect.ValueOf(t).Convert(dst.Type()))
			return nil
		}

		if dst.Type() == netipAddrType {
			a, err := netip.ParseAddr(pv.String())
			if err != nil {
				return fmt.Errorf("invalid %s value %q: %w", dst.Type(), pv.String(), err)
			}
			dst.Set(reflect.ValueOf(a))
			return nil
		}

		if dst.Type() == uuidType {
			u, err := uuid.Parse(pv.String())
			if err != nil {
				return fmt.Errorf("invalid %s value %q: %w", dst.Type(), pv.String(), err)
			}
			dst.Set(reflect.ValueOf(u))
			return nil
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
	return nil
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

			// do not allocate a missing embedded model pointer; leave its proto fields unset
			field, ok := fieldByIndex(mv, fm.idx, false)
			if !ok {
				continue
			}
			goToPbField(msg, fm.fd, field)
		}
		return out
	}
}

// NewFromPb returns a func converting a proto input message to a store input struct. Fields are matched by proto field name = input json name.
// The returned func errors if a string value cannot be parsed as its Go field type. NewFromPb panics if the proto fields
// and the input fields are not an exact match or have unsupported types, so call it at init.
func NewFromPb[inputT any, pbT any, pbPtrT interface {
	*pbT
	proto.Message
}]() func(pbPtrT) (inputT, error) {

	desc := pbPtrT(new(pbT)).ProtoReflect().Descriptor()
	maps := buildPbFieldMaps(desc, reflect.TypeFor[inputT](), true)

	return func(p pbPtrT) (inputT, error) {
		var in inputT
		iv := reflect.ValueOf(&in).Elem()
		msg := p.ProtoReflect()
		for _, fm := range maps {

			// allocate embedded pointers in the output struct as needed
			field, ok := fieldByIndex(iv, fm.idx, true)
			if !ok {
				panic(fmt.Sprintf("cannot set input field %q", fm.fd.Name()))
			}
			if err := pbToGoField(msg, fm.fd, field); err != nil {
				return in, err
			}
		}
		return in, nil
	}
}
