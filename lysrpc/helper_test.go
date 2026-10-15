package lysrpc

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/loveyourstack/lys/internal/cmd"
	"github.com/loveyourstack/lys/internal/myapp"
	"github.com/loveyourstack/lys/internal/pb"
	"github.com/loveyourstack/lys/internal/stores/core/coreparamtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/internal/stores/core/coreuuidtest"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lyslog"
	"github.com/loveyourstack/lys/lyspgdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

type gRpcServerApplication struct {
	*cmd.Application
	GetOptions lysapi.GetOptions
}

func mustGetRpcSrvApp(ctx context.Context, t testing.TB) *gRpcServerApplication {

	conf := myapp.MustGetConfig(t)

	app := &cmd.Application{
		Config:   &conf,
		Logger:   slog.New(lyslog.NewSplitStreamHandler(os.Stdout, os.Stderr, nil)),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	// create gRPC server app
	srvApp := &gRpcServerApplication{
		Application: app,
	}

	var err error
	srvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{})
	if err != nil {
		t.Fatalf("FillGetOptions failed: %v", err)
	}

	// register core.weekday type in any conn added to the pool so that Patch of type_test core.weekday[] works. If don't do this: "encode plan not found"
	dataTypeNames := []string{
		"core.weekday",
		"core.weekday[]",
	}
	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbOwnerUser, "test", dataTypeNames)
	if err != nil {
		t.Fatalf("lyspgdb.GetPoolWithTypes failed: %v", err)
	}

	return srvApp
}

// core.type_test converters: on compilation, these will check that all types are handled correctly in lysrpc meta funcs

var CoreTTModelToPbFunc = NewToPb[coretypetestm.Model, pb.CoreTypeTest]()
var CoreTTInputToPbFunc = NewToPb[coretypetestm.Input, pb.CoreTypeTestInput]()

var CoreTTModelFromPbFunc = NewFromPb[coretypetestm.Model, pb.CoreTypeTest]()
var CoreTTInputFromPbFunc = NewFromPb[coretypetestm.Input, pb.CoreTypeTestInput]()

type coreTypeTestServer struct {
	pb.UnimplementedCoreTypeTestSvcServer
	*gRpcServerApplication

	store            coretypetest.Store
	jsonKeyDbNameMap map[string]string
}

func (s *coreTypeTestServer) Create(ctx context.Context, req *pb.CoreTypeTestInput) (*pb.CoreTypeTest, error) {
	return Create(ctx, req, s.store, CoreTTInputFromPbFunc, CoreTTModelToPbFunc)
}

func (s *coreTypeTestServer) Delete(ctx context.Context, req *pb.DeleteByIdReq) (*emptypb.Empty, error) {
	return Delete(ctx, req.Id, s.store)
}

func (s *coreTypeTestServer) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.CoreTypeTest, error) {
	return GetById(ctx, req.Id, s.store, CoreTTModelToPbFunc)
}

func (s *coreTypeTestServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListCoreTypeTestResp, error) {
	return List(ctx, s.store, req, s.GetOptions, CoreTTModelToPbFunc,
		func(data []*pb.CoreTypeTest, meta lysapi.SelectMetadata) *pb.ListCoreTypeTestResp {
			return &pb.ListCoreTypeTestResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *coreTypeTestServer) Update(ctx context.Context, req *pb.UpdateCoreTypeTestReq) (*pb.CoreTypeTest, error) {
	return Update(ctx, req.Input, req.FieldMask, req.Id, s.jsonKeyDbNameMap, s.store, CoreTTModelToPbFunc)
}

var CoreUTModelToPbFunc = NewToPb[coreuuidtest.Model, pb.CoreUuidTest]()
var CoreUTInputToPbFunc = NewToPb[coreuuidtest.Input, pb.CoreUuidTestInput]()

var CoreUTModelFromPbFunc = NewFromPb[coreuuidtest.Model, pb.CoreUuidTest]()
var CoreUTInputFromPbFunc = NewFromPb[coreuuidtest.Input, pb.CoreUuidTestInput]()

type coreUuidTestServer struct {
	pb.UnimplementedCoreUuidTestSvcServer
	*gRpcServerApplication

	store            coreuuidtest.Store
	jsonKeyDbNameMap map[string]string
}

func (s *coreUuidTestServer) Create(ctx context.Context, req *pb.CoreUuidTestInput) (*pb.CoreUuidTest, error) {
	return Create(ctx, req, s.store, CoreUTInputFromPbFunc, CoreUTModelToPbFunc)
}

func (s *coreUuidTestServer) Delete(ctx context.Context, req *pb.DeleteByUuidReq) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid")
	}
	return Delete(ctx, id, s.store)
}

func (s *coreUuidTestServer) Get(ctx context.Context, req *pb.GetByUuidReq) (*pb.CoreUuidTest, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid")
	}
	return GetById(ctx, id, s.store, CoreUTModelToPbFunc)
}

func (s *coreUuidTestServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListCoreUuidTestResp, error) {
	return List(ctx, s.store, req, s.GetOptions, CoreUTModelToPbFunc,
		func(data []*pb.CoreUuidTest, meta lysapi.SelectMetadata) *pb.ListCoreUuidTestResp {
			return &pb.ListCoreUuidTestResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

func (s *coreUuidTestServer) Update(ctx context.Context, req *pb.UpdateCoreUuidTestReq) (*pb.CoreUuidTest, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid uuid")
	}
	return Update(ctx, req.Input, req.FieldMask, id, s.jsonKeyDbNameMap, s.store, CoreUTModelToPbFunc)
}

var CorePTModelToPbFunc = NewToPb[coreparamtest.Model, pb.CoreParamTest]()
var CorePTInputToPbFunc = NewToPb[coreparamtest.Input, pb.CoreParamTestInput]()

var CorePTModelFromPbFunc = NewFromPb[coreparamtest.Model, pb.CoreParamTest]()
var CorePTInputFromPbFunc = NewFromPb[coreparamtest.Input, pb.CoreParamTestInput]()

type coreParamTestServer struct {
	pb.UnimplementedCoreParamTestSvcServer
	*gRpcServerApplication

	store            coreparamtest.Store
	jsonKeyDbNameMap map[string]string
}

func (s *coreParamTestServer) Get(ctx context.Context, req *pb.GetByIdReq) (*pb.CoreParamTest, error) {
	return GetById(ctx, req.Id, s.store, CorePTModelToPbFunc)
}

func (s *coreParamTestServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListCoreParamTestResp, error) {
	return List(ctx, s.store, req, s.GetOptions, CorePTModelToPbFunc,
		func(data []*pb.CoreParamTest, meta lysapi.SelectMetadata) *pb.ListCoreParamTestResp {
			return &pb.ListCoreParamTestResp{Data: data, Metadata: toPbMetadata(meta, len(data))}
		})
}

// mustGetRpcConn creates and returns a gRPC ClientConn connected to an in-memory gRPC server for testing.
func mustGetRpcConn(ctx context.Context, t testing.TB) *grpc.ClientConn {

	app := mustGetRpcSrvApp(ctx, t)

	// create in-memory listener for the gRPC server
	lis := bufconn.Listen(1 << 20)

	// create gRPC server
	srv := grpc.NewServer()

	// register services

	ttStore := coretypetest.Store{Db: app.Db}
	pb.RegisterCoreTypeTestSvcServer(srv, &coreTypeTestServer{gRpcServerApplication: app, store: ttStore,
		jsonKeyDbNameMap: ttStore.GetPlan().JsonKeyDbNameMap(),
	})

	utStore := coreuuidtest.Store{Db: app.Db}
	pb.RegisterCoreUuidTestSvcServer(srv, &coreUuidTestServer{gRpcServerApplication: app, store: utStore,
		jsonKeyDbNameMap: utStore.GetPlan().JsonKeyDbNameMap(),
	})

	ptStore := coreparamtest.Store{Db: app.Db}
	pb.RegisterCoreParamTestSvcServer(srv, &coreParamTestServer{gRpcServerApplication: app, store: ptStore,
		jsonKeyDbNameMap: ptStore.GetPlan().JsonKeyDbNameMap(),
	})

	go srv.Serve(lis)

	t.Cleanup(app.Db.Close)
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient failed: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

func toPbMetadata(m lysapi.SelectMetadata, itemCount int) *pb.GetMetadata {
	return &pb.GetMetadata{
		Count:                 int32(itemCount),
		TotalCount:            m.TotalCount,
		TotalCountIsEstimated: m.TotalCountIsEstimated,
	}
}
