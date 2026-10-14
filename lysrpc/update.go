package lysrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Update performs a partial update on a model identified by its ID using the provided field mask and input message. It returns the updated protobuf model.
func Update[idT lyspg.PrimaryKeyType, inputT proto.Message, modelT any, pbModelT any](
	ctx context.Context,
	input inputT,
	fieldMask *fieldmaskpb.FieldMask,
	id idT,
	jsonKeyDbNameMap map[string]string,
	store lyspg.UpdatePartialSelecter[idT, modelT],
	toPb func(modelT) *pbModelT,
) (*pbModelT, error) {

	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return nil, status.Error(codes.InvalidArgument, "field mask is required")
	}

	// build the assignment map based on the field mask and input
	assMap, err := buildRpcUpdateAssignmentMap(input, fieldMask, jsonKeyDbNameMap)
	if err != nil {
		return nil, err
	}

	// perform the partial update in the store
	err = store.UpdatePartial(ctx, assMap, id)
	if err != nil {
		if userErr, ok := errors.AsType[lyserr.User](err); ok {
			return nil, status.Error(codes.InvalidArgument, userErr.Error())
		}
		return nil, fmt.Errorf("store.UpdatePartial failed: %w", err)
	}

	// retrieve the updated item from the store
	item, err := store.SelectById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, fmt.Errorf("store.SelectById failed: %w", err)
	}

	return toPb(item), nil
}

func buildRpcUpdateAssignmentMap(input proto.Message, fieldMask *fieldmaskpb.FieldMask, jsonKeyDbNameMap map[string]string) (assMap map[string]any, err error) {

	// reflect input and ensure it is present
	msg := input.ProtoReflect()
	if !msg.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "input is required")
	}

	// check for unknown fields in the input message (only works using binary wire format)
	if len(msg.GetUnknown()) > 0 {
		return nil, status.Error(codes.InvalidArgument, "input contains unknown fields")
	}

	// normalize the field mask to ensure all paths are in canonical form
	fieldMask.Normalize()

	// check field mask against input message fields
	if !fieldMask.IsValid(input) {
		return nil, status.Error(codes.InvalidArgument, "invalid field mask")
	}

	fields := msg.Descriptor().Fields()

	assMap = make(map[string]any, len(fieldMask.Paths))

	// for each path (field)
	for _, p := range fieldMask.Paths {

		// ensure it is valid
		if _, ok := jsonKeyDbNameMap[p]; !ok {
			return nil, status.Errorf(codes.InvalidArgument, "invalid field mask path: %s", p)
		}

		// get the field descriptor
		fd := fields.ByName(protoreflect.Name(p))
		if fd == nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown field: %s", p)
		}

		// if an optional field is in the mask but has no input value, set it to nil
		if fd.HasPresence() && !msg.Has(fd) {
			assMap[p] = nil
			continue
		}

		// get the reflected value and add it as an assignment
		v, err := protoValueToAny(fd, msg.Get(fd))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "field %s: %s", p, err.Error())
		}

		assMap[p] = v
	}

	return assMap, nil
}

// protoValueToAny converts a protobuf field value to a Go any type, handling repeated fields appropriately.
func protoValueToAny(fd protoreflect.FieldDescriptor, v protoreflect.Value) (any, error) {

	if fd.IsMap() {
		return nil, errors.New("map fields are not supported")
	}

	// if the field is not a list, just return its value as-is
	if !fd.IsList() {
		return v.Interface(), nil
	}

	l := v.List()

	switch fd.Kind() {
	case protoreflect.StringKind:
		return protoListToSlice(l, protoreflect.Value.String), nil
	case protoreflect.BoolKind:
		return protoListToSlice(l, protoreflect.Value.Bool), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoListToSlice(l, func(v protoreflect.Value) int32 { return int32(v.Int()) }), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoListToSlice(l, protoreflect.Value.Int), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoListToSlice(l, func(v protoreflect.Value) int64 { return int64(v.Uint()) }), nil // postgres has no unsigned ints
	case protoreflect.FloatKind:
		return protoListToSlice(l, func(v protoreflect.Value) float32 { return float32(v.Float()) }), nil
	case protoreflect.DoubleKind:
		return protoListToSlice(l, protoreflect.Value.Float), nil
	case protoreflect.EnumKind:
		return protoListToSlice(l, func(v protoreflect.Value) string {
			if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
				return string(ev.Name())
			}
			return ""
		}), nil
	default:
		return nil, fmt.Errorf("repeated %s fields are not supported", fd.Kind())
	}
}

// protoListToSlice always returns a non-nil slice, so an empty list is written as '{}' rather than NULL
func protoListToSlice[T any](l protoreflect.List, conv func(protoreflect.Value) T) []T {
	out := make([]T, l.Len())
	for i := range out {
		out[i] = conv(l.Get(i))
	}
	return out
}
