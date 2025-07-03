package functions

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/Nu11ified/golem/proto/gen/proto"
)

// GRPCServer implements the FunctionService gRPC interface
type GRPCServer struct {
	pb.UnimplementedFunctionServiceServer
	registry *Registry
}

// NewGRPCServer creates a new gRPC server with the function registry
func NewGRPCServer(registry *Registry) *GRPCServer {
	return &GRPCServer{
		registry: registry,
	}
}

// Call implements the Call RPC method
func (s *GRPCServer) Call(ctx context.Context, req *pb.FunctionRequest) (*pb.FunctionResponse, error) {
	log.Printf("gRPC Call: %s.%s with %d args", req.ServiceName, req.FunctionName, len(req.Args))

	// Call the function through the registry
	result, err := s.registry.CallFunction(ctx, req.ServiceName, req.FunctionName, req.Args)
	if err != nil {
		log.Printf("Function call error: %v", err)
		return &pb.FunctionResponse{
			Success:  false,
			Error:    err.Error(),
			Metadata: make(map[string]string),
		}, nil
	}

	return &pb.FunctionResponse{
		Success:  true,
		Result:   result,
		Metadata: make(map[string]string),
	}, nil
}

// CallStream implements the streaming Call RPC method
func (s *GRPCServer) CallStream(req *pb.FunctionRequest, stream pb.FunctionService_CallStreamServer) error {
	// For now, just call the function once and send the result
	// This could be extended for true streaming functionality
	ctx := stream.Context()

	result, err := s.registry.CallFunction(ctx, req.ServiceName, req.FunctionName, req.Args)
	if err != nil {
		return stream.Send(&pb.FunctionResponse{
			Success:  false,
			Error:    err.Error(),
			Metadata: make(map[string]string),
		})
	}

	return stream.Send(&pb.FunctionResponse{
		Success:  true,
		Result:   result,
		Metadata: make(map[string]string),
	})
}

// ListFunctions implements the ListFunctions RPC method
func (s *GRPCServer) ListFunctions(ctx context.Context, req *pb.ListFunctionsRequest) (*pb.ListFunctionsResponse, error) {
	log.Printf("gRPC ListFunctions for service: %s", req.ServiceName)

	functions := s.registry.ListFunctions(req.ServiceName)

	return &pb.ListFunctionsResponse{
		Functions: functions,
	}, nil
}

// HTTPHandler provides a simple HTTP endpoint for testing
func (s *GRPCServer) HTTPHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle CORS preflight
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		// Parse request
		var reqData struct {
			FunctionName string        `json:"functionName"`
			ServiceName  string        `json:"serviceName"`
			Args         []interface{} `json:"args"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}

		// Convert args to protobuf Any
		var protoArgs []*anypb.Any
		for _, arg := range reqData.Args {
			argBytes, err := json.Marshal(arg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to serialize argument"})
				return
			}

			anyArg := &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Value",
				Value:   argBytes,
			}
			protoArgs = append(protoArgs, anyArg)
		}

		// Call function
		result, err := s.registry.CallFunction(r.Context(), reqData.ServiceName, reqData.FunctionName, protoArgs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Convert result back to JSON
		resultBytes := result.GetValue()
		var resultData interface{}
		if err := json.Unmarshal(resultBytes, &resultData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to deserialize result"})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  resultData,
		})
	}
}

// CreateGRPCServer creates and configures a gRPC server
func CreateGRPCServer(registry *Registry) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	functionServer := NewGRPCServer(registry)
	pb.RegisterFunctionServiceServer(grpcServer, functionServer)

	return grpcServer
}

// loggingInterceptor logs all gRPC calls
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC call: %s", info.FullMethod)

	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("gRPC call %s failed: %v", info.FullMethod, err)
	} else {
		log.Printf("gRPC call %s succeeded", info.FullMethod)
	}

	return resp, err
}
