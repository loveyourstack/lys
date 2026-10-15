package lysrpc

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/loveyourstack/lys/internal/pb"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestFieldByIndex(t *testing.T) {
	type Leaf struct {
		Value int
	}
	type holder struct {
		*Leaf
	}

	var h holder
	field, ok := fieldByIndex(reflect.ValueOf(&h).Elem(), []int{0, 0}, false)
	assert.False(t, ok)
	assert.False(t, field.IsValid())
	assert.Nil(t, h.Leaf)

	field, ok = fieldByIndex(reflect.ValueOf(&h).Elem(), []int{0, 0}, true)
	require.True(t, ok)
	field.SetInt(42)
	assert.Equal(t, 42, h.Value)

	field, ok = fieldByIndex(reflect.ValueOf(h), []int{0, 0}, false)
	require.True(t, ok)
	assert.Equal(t, int64(42), field.Int())

	field, ok = fieldByIndex(reflect.ValueOf(&h).Elem(), []int{0, 0, 0}, true)
	assert.False(t, ok)
	assert.False(t, field.IsValid())
}

func TestJsonFieldIndexes(t *testing.T) {
	type embedded struct {
		Embedded string
	}
	type fields struct {
		embedded
		Tagged  int `json:"tagged,omitempty"`
		Default bool
		Ignored string `json:"-"`
		hidden  string
	}

	indexes := jsonFieldIndexes(reflect.TypeFor[fields]())
	assert.Equal(t, []int{0, 0}, indexes["Embedded"])
	assert.Equal(t, []int{1}, indexes["tagged"])
	assert.Equal(t, []int{2}, indexes["Default"])
	assert.NotContains(t, indexes, "embedded")
	assert.NotContains(t, indexes, "Ignored")
	assert.NotContains(t, indexes, "hidden")
}

func TestIsTimestamp(t *testing.T) {
	desc := (&pb.CoreTypeTestInput{}).ProtoReflect().Descriptor()
	timestamp := desc.Fields().ByName("c_datetime")
	text := desc.Fields().ByName("c_date")

	assert.True(t, isTimestamp(timestamp))
	assert.False(t, isTimestamp(text))
}

func TestBuildPbFieldMaps(t *testing.T) {
	desc := (&pb.CoreTypeTestInput{}).ProtoReflect().Descriptor()
	maps := buildPbFieldMaps(desc, reflect.TypeFor[coretypetestm.Input](), true)
	assert.Len(t, maps, desc.Fields().Len())

	dateMap := assertPbFieldMap(t, maps, "c_date")
	assert.Equal(t, reflect.TypeFor[lystype.Date](), reflect.TypeFor[coretypetestm.Input]().FieldByIndex(dateMap.idx).Type)

	assert.Panics(t, func() {
		buildPbFieldMaps(desc, reflect.TypeFor[struct {
			Extra string
		}](), false)
	})
	assert.Panics(t, func() {
		buildPbFieldMaps(desc, reflect.TypeFor[struct {
			coretypetestm.Input
			Extra string
		}](), true)
	})
}

func assertPbFieldMap(t *testing.T, maps []pbFieldMap, name string) pbFieldMap {
	t.Helper()
	for _, fieldMap := range maps {
		if string(fieldMap.fd.Name()) == name {
			return fieldMap
		}
	}
	t.Fatalf("field map %q not found", name)
	return pbFieldMap{}
}

func TestIntegerKinds(t *testing.T) {
	for _, kind := range []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64} {
		assert.True(t, isSignedInteger(kind), kind.String())
		assert.False(t, isUnsignedInteger(kind), kind.String())
	}
	for _, kind := range []reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64} {
		assert.True(t, isUnsignedInteger(kind), kind.String())
		assert.False(t, isSignedInteger(kind), kind.String())
	}
	for _, kind := range []reflect.Kind{reflect.Bool, reflect.Float32, reflect.Float64, reflect.String} {
		assert.False(t, isSignedInteger(kind), kind.String())
		assert.False(t, isUnsignedInteger(kind), kind.String())
	}
}

func TestCheckFieldCompat(t *testing.T) {
	fields := (&pb.CoreTypeTestInput{}).ProtoReflect().Descriptor().Fields()
	tests := []struct {
		name string
		fd   protoreflect.FieldDescriptor
		typ  reflect.Type
		ok   bool
	}{
		{"bool", fields.ByName("c_bool"), reflect.TypeFor[bool](), true},
		{"nullable integer", fields.ByName("c_intn"), reflect.TypeFor[*int64](), true},
		{"integer width mismatch", fields.ByName("c_int"), reflect.TypeFor[int32](), false},
		{"repeated", fields.ByName("c_inta"), reflect.TypeFor[[]int64](), true},
		{"repeated pointers", fields.ByName("c_inta"), reflect.TypeFor[[]*int64](), false},
		{"date layout", fields.ByName("c_date"), reflect.TypeFor[lystype.Date](), true},
		{"string type mismatch", fields.ByName("c_text"), reflect.TypeFor[int64](), false},
		{"timestamp", fields.ByName("c_datetime"), reflect.TypeFor[time.Time](), true},
		{"timestamp pointer", fields.ByName("c_datetimen"), reflect.TypeFor[*time.Time](), true},
		{"timestamp mismatch", fields.ByName("c_datetime"), reflect.TypeFor[int64](), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkFieldCompat(tt.fd, tt.typ)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEnumValueConversion(t *testing.T) {
	enumName := "Weekday"
	enumValue0, enumValue1 := "UNSPECIFIED", "MONDAY"
	fieldName := "weekday"
	syntax := "proto3"
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:    new("meta_test.proto"),
		Package: new("meta_test"),
		Syntax:  &syntax,
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: &enumName,
			Value: []*descriptorpb.EnumValueDescriptorProto{
				{Name: &enumValue0, Number: new(int32(0))},
				{Name: &enumValue1, Number: new(int32(1))},
			},
		}},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: new("Entry"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     &fieldName,
				Number:   new(int32(1)),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
				TypeName: new(".meta_test.Weekday"),
			}},
		}},
	}, nil)
	require.NoError(t, err)
	fd := file.Messages().ByName("Entry").Fields().ByName("weekday")

	pv, ok := goToPbValue(fd, reflect.ValueOf("MONDAY"))
	require.True(t, ok)
	assert.Equal(t, protoreflect.EnumNumber(1), pv.Enum())
	_, ok = goToPbValue(fd, reflect.ValueOf("NOT_A_DAY"))
	assert.False(t, ok)

	msg := dynamicpb.NewMessage(file.Messages().ByName("Entry"))
	msg.Set(fd, pv)
	dst := reflect.New(reflect.TypeFor[string]()).Elem()
	pbToGoValue(fd, msg.Get(fd), dst)
	assert.Equal(t, "MONDAY", dst.String())

	dst.SetString("unchanged")
	pbToGoValue(fd, protoreflect.ValueOfEnum(99), dst)
	assert.Equal(t, "unchanged", dst.String())
}

func TestNewToPbAndNewFromPb(t *testing.T) {
	input, err := coretypetestm.GetFilledInput()
	require.NoError(t, err)

	toPb := NewToPb[coretypetestm.Input, pb.CoreTypeTestInput]()
	fromPb := NewFromPb[coretypetestm.Input, pb.CoreTypeTestInput]()
	message := toPb(input)
	require.NotNil(t, message)

	assert.Equal(t, input.CBool, message.CBool)
	assert.Equal(t, input.CInt, message.CInt)
	assert.Equal(t, input.CDate.Format(lystype.DateFormat), message.CDate)
	assert.Equal(t, input.CTime.Format(lystype.TimeFormat), message.CTime)
	require.NotNil(t, message.CDatetime)
	assert.True(t, time.Time(input.CDatetime).Equal(message.CDatetime.AsTime()))
	assert.Equal(t, input.CIntA, message.CInta)
	assert.NotNil(t, message.CBooln)
	assert.Equal(t, *input.CBoolN, *message.CBooln)

	output := fromPb(message)
	assert.Equal(t, input.CBool, output.CBool)
	assert.Equal(t, input.CBoolN, output.CBoolN)
	assert.Equal(t, input.CBoolA, output.CBoolA)
	assert.Equal(t, input.CInt, output.CInt)
	assert.Equal(t, input.CIntN, output.CIntN)
	assert.Equal(t, input.CIntA, output.CIntA)
	assert.Equal(t, input.CDouble, output.CDouble)
	assert.Equal(t, input.CDoubleN, output.CDoubleN)
	assert.Equal(t, input.CDoubleA, output.CDoubleA)
	assert.Equal(t, input.CNumeric, output.CNumeric)
	assert.Equal(t, input.CNumericN, output.CNumericN)
	assert.Equal(t, input.CNumericA, output.CNumericA)
	assert.Equal(t, input.CDate.Format(lystype.DateFormat), output.CDate.Format(lystype.DateFormat))
	assert.Equal(t, input.CDateN.Format(lystype.DateFormat), output.CDateN.Format(lystype.DateFormat))
	assert.Equal(t, input.CDateA, output.CDateA)
	assert.Equal(t, input.CTime.Format(lystype.TimeFormat), output.CTime.Format(lystype.TimeFormat))
	assert.Equal(t, input.CTimeN.Format(lystype.TimeFormat), output.CTimeN.Format(lystype.TimeFormat))
	assert.Equal(t, input.CTimeA, output.CTimeA)
	assert.True(t, time.Time(input.CDatetime).Equal(time.Time(output.CDatetime)))
	assert.True(t, time.Time(*input.CDatetimeN).Equal(time.Time(*output.CDatetimeN)))
	require.Len(t, output.CDatetimeA, len(input.CDatetimeA))
	for i := range input.CDatetimeA {
		assert.True(t, time.Time(input.CDatetimeA[i]).Equal(time.Time(output.CDatetimeA[i])))
	}
	assert.Equal(t, input.CEnum, output.CEnum)
	assert.Equal(t, input.CEnumN, output.CEnumN)
	assert.Equal(t, input.CEnumA, output.CEnumA)
	assert.Equal(t, input.CText, output.CText)
	assert.Equal(t, input.CTextN, output.CTextN)
	assert.Equal(t, input.CTextA, output.CTextA)
}

func TestNewToPbZeroValues(t *testing.T) {
	input := coretypetestm.GetEmptyInput()
	message := NewToPb[coretypetestm.Input, pb.CoreTypeTestInput]()(input)

	assert.Nil(t, message.CBooln)
	assert.Nil(t, message.CIntn)
	assert.Nil(t, message.CDatetime)
	assert.Nil(t, message.CDatetimen)
	assert.Empty(t, message.CDatetimea)
	assert.Empty(t, message.CInta)

	output := NewFromPb[coretypetestm.Input, pb.CoreTypeTestInput]()(&pb.CoreTypeTestInput{})
	assert.Nil(t, output.CBoolN)
	assert.Nil(t, output.CDatetimeN)
	assert.NotNil(t, output.CBoolA)
	assert.Empty(t, output.CBoolA)
	assert.NotNil(t, output.CDatetimeA)
	assert.Empty(t, output.CDatetimeA)
}

func TestNewToPbUuid(t *testing.T) {
	toPb := NewToPb[coretypetestm.Model, pb.CoreTypeTest]()
	assert.Empty(t, toPb(coretypetestm.Model{}).IdUu)

	id := uuid.MustParse("a4ca421a-912f-4ac8-b3e3-c97b557dad1d")
	assert.Equal(t, id.String(), toPb(coretypetestm.Model{Iduu: id}).IdUu)
}

func TestNewConvertersEmbeddedPointers(t *testing.T) {
	type inputWithEmbeddedPointer struct {
		*coretypetestm.Input
	}

	toPb := NewToPb[inputWithEmbeddedPointer, pb.CoreTypeTestInput]()
	assert.Empty(t, toPb(inputWithEmbeddedPointer{}).CText)

	input, err := coretypetestm.GetFilledInput()
	require.NoError(t, err)
	message := NewToPb[coretypetestm.Input, pb.CoreTypeTestInput]()(input)

	fromPb := NewFromPb[inputWithEmbeddedPointer, pb.CoreTypeTestInput]()
	output := fromPb(message)
	require.NotNil(t, output.Input)
	assert.Equal(t, input.CText, output.CText)
	assert.Equal(t, input.CTextN, output.CTextN)
}
