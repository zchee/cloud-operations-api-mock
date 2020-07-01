// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloud-operations-api-mock/internal/validation"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	st "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

const bufSize = 1024 * 1024

var (
	client     monitoring.MetricServiceClient
	conn       *grpc.ClientConn
	ctx        context.Context
	grpcServer *grpc.Server
	lis        *bufconn.Listener
)

func setup() {
	// Setup the in-memory server.
	lis = bufconn.Listen(bufSize)
	grpcServer = grpc.NewServer()
	monitoring.RegisterMetricServiceServer(grpcServer, NewMockMetricServer())
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()

	// Setup the connection and client.
	ctx = context.Background()
	var err error
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial bufnet: %v", err)
	}
	client = monitoring.NewMetricServiceClient(conn)
}

func tearDown() {
	conn.Close()
	grpcServer.GracefulStop()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestMockMetricServer_CreateTimeSeries(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.CreateTimeSeriesRequest{
		Name:       "test create time series request",
		TimeSeries: []*monitoring.TimeSeries{&monitoring.TimeSeries{}},
	}
	want := &empty.Empty{}
	response, err := client.CreateTimeSeries(ctx, in)
	if err != nil {
		t.Fatalf("failed to call CreateTimeSeries %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("CreateTimeSeries(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_ListTimeSeries(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListTimeSeriesRequest{
		Name:     "test list time series request",
		Filter:   "test filter",
		Interval: &monitoring.TimeInterval{},
		View:     monitoring.ListTimeSeriesRequest_HEADERS,
	}
	want := &monitoring.ListTimeSeriesResponse{
		TimeSeries:      []*monitoring.TimeSeries{},
		NextPageToken:   "",
		ExecutionErrors: []*status.Status{},
	}

	response, err := client.ListTimeSeries(ctx, in)
	if err != nil {
		t.Fatalf("failed to call ListTimeSeries %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("ListTimeSeries(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_GetMonitoredResourceDescriptor(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.GetMonitoredResourceDescriptorRequest{
		Name: "test get metric monitored resource descriptor",
	}
	want := &monitoredres.MonitoredResourceDescriptor{}
	response, err := client.GetMonitoredResourceDescriptor(ctx, in)
	if err != nil {
		t.Fatalf("failed to call GetMonitoredResourceDescriptor %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("GetMonitoredResourceDescriptor(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_ListMonitoredResourceDescriptors(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListMonitoredResourceDescriptorsRequest{
		Name: "test list monitored resource descriptors",
	}
	want := &monitoring.ListMonitoredResourceDescriptorsResponse{
		ResourceDescriptors: []*monitoredres.MonitoredResourceDescriptor{},
	}
	response, err := client.ListMonitoredResourceDescriptors(ctx, in)
	if err != nil {
		t.Fatalf("failed to call ListMonitoredResourceDescriptors %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("ListMonitoredResourceDescriptors(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_GetMetricDescriptor(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.GetMetricDescriptorRequest{
		Name: "test-metric-descriptor-1",
	}
	want := &metric.MetricDescriptor{
		Name: "test-metric-descriptor-1",
	}

	if _, err := client.CreateMetricDescriptor(ctx, &monitoring.CreateMetricDescriptorRequest{
		Name: "test-metric-descriptor-1",
		MetricDescriptor: &metric.MetricDescriptor{
			Name: "test-metric-descriptor-1",
		},
	}); err != nil {
		t.Fatalf("failed to create test metric descriptor with error: %v", err)
	}

	response, err := client.GetMetricDescriptor(ctx, in)
	if err != nil {
		t.Fatalf("failed to call GetMetricDescriptor %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("GetMetricDescriptor(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_CreateMetricDescriptor(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.CreateMetricDescriptorRequest{
		Name:             "test create metric descriptor",
		MetricDescriptor: &metric.MetricDescriptor{},
	}
	want := &metric.MetricDescriptor{}
	response, err := client.CreateMetricDescriptor(ctx, in)
	if err != nil {
		t.Fatalf("failed to call CreateMetricDescriptor: %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("CreateMetricDescriptor(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_DeleteMetricDescriptor(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.DeleteMetricDescriptorRequest{
		Name: "test-metric-descriptor",
	}
	want := &empty.Empty{}

	if _, err := client.CreateMetricDescriptor(ctx, &monitoring.CreateMetricDescriptorRequest{
		Name: "test",
		MetricDescriptor: &metric.MetricDescriptor{
			Name: "test-metric-descriptor",
		},
	}); err != nil {
		t.Fatalf("failed to create test metric descriptor with error: %v", err)
	}

	response, err := client.DeleteMetricDescriptor(ctx, in)
	if err != nil {
		t.Fatalf("failed to call DeleteMetricDescriptorRequest: %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("DeleteMetricDescriptorRequest(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_ListMetricDescriptors(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListMetricDescriptorsRequest{
		Name: "test list metric decriptors request",
	}
	want := &monitoring.ListMetricDescriptorsResponse{
		MetricDescriptors: []*metric.MetricDescriptor{},
	}
	response, err := client.ListMetricDescriptors(ctx, in)
	if err != nil {
		t.Fatalf("failed to call ListMetricDescriptors %v", err)
	}

	if !proto.Equal(response, want) {
		t.Errorf("ListMetricDescriptors(%q) == %q, want %q", in, response, want)
	}
}

func TestMockMetricServer_GetMetricDescriptor_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.GetMetricDescriptorRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}}
	response, err := client.GetMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("GetMetricDescriptor(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("GetMetricDescriptor(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_GetMonitoredResourceDescriptor_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.GetMonitoredResourceDescriptorRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}}
	response, err := client.GetMonitoredResourceDescriptor(ctx, in)
	if err == nil {
		t.Errorf("GetMonitoredResourceDescriptor(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("GetMonitoredResourceDescriptor(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_GetMetricDescriptor_NotFoundError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.GetMetricDescriptorRequest{
		Name: "test",
	}
	want := validation.StatusMetricDescriptorNotFound
	response, err := client.GetMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("GetMetricDescriptor(%q) == %q, expected error %q", in, response, want.Message())
	}

	if s := st.Convert(err); s.Code() != want.Code() {
		t.Errorf("GetMetricDescriptor(%q) returned error %q, expected error %q",
			in, s.Message(), want.Message())
	}
}

func TestMockMetricServer_DeleteMetricDescriptor_NotFoundError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.DeleteMetricDescriptorRequest{
		Name: "test",
	}
	want := validation.StatusMetricDescriptorNotFound
	response, err := client.DeleteMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("DeleteMetricDescriptor(%q) == %q, expected error %q", in, response, want.Message())
	}

	if s := st.Convert(err); s.Code() != want.Code() {
		t.Errorf("DeleteMetricDescriptor(%q) returned error %q, expected error %q",
			in, s.Message(), want.Message())
	}
}

func TestMockMetricServer_MetricDescriptor_DataRace(t *testing.T) {
	setup()
	defer tearDown()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := client.CreateMetricDescriptor(ctx, &monitoring.CreateMetricDescriptorRequest{
			Name:             "test-create-metric-descriptor",
			MetricDescriptor: &metric.MetricDescriptor{Name: "test-metric-descriptor-1"},
		})
		if err != nil {
			t.Fatalf("failed to call CreateMetricDescriptor: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := client.CreateMetricDescriptor(ctx, &monitoring.CreateMetricDescriptorRequest{
			Name:             "test-create-metric-descriptor",
			MetricDescriptor: &metric.MetricDescriptor{Name: "test-metric-descriptor-2"},
		})
		if err != nil {
			t.Fatalf("failed to call CreateMetricDescriptor: %v", err)
		}
	}()

	wg.Wait()
}

func TestMockMetricServer_DuplicateMetricDescriptorError(t *testing.T) {
	setup()
	defer tearDown()
	duplicateSpanName := "test-metric-descriptor-1"

	if _, err := client.CreateMetricDescriptor(ctx, &monitoring.CreateMetricDescriptorRequest{
		Name: "test",
		MetricDescriptor: &metric.MetricDescriptor{
			Name: duplicateSpanName,
		},
	}); err != nil {
		t.Fatalf("failed to create test metric descriptor with error: %v", err)
	}

	in := &monitoring.CreateMetricDescriptorRequest{
		Name: "test",
		MetricDescriptor: &metric.MetricDescriptor{
			Name: duplicateSpanName,
		},
	}
	want := validation.StatusDuplicateMetricDescriptorName
	response, err := client.CreateMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("CreateMetricDescriptor(%q) == %q, expected error %q", in, response, want.Message())
	}

	if valid := validation.ValidateDuplicateSpanNames(err, duplicateSpanName); !valid {
		t.Errorf("expected duplicate spanName: %v", duplicateSpanName)
	}

}

func TestMockMetricServer_DeleteMetricDescriptor_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.DeleteMetricDescriptorRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}}
	response, err := client.DeleteMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("DeleteMetricDescriptor(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("DeleteMetricDescriptor(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_ListMetricDescriptor_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListMetricDescriptorsRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}}
	response, err := client.ListMetricDescriptors(ctx, in)
	if err == nil {
		t.Errorf("ListMetricDescriptors(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("ListMetricDescriptors(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_CreateMetricDescriptor_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.CreateMetricDescriptorRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}, "MetricDescriptor": {}}
	response, err := client.CreateMetricDescriptor(ctx, in)
	if err == nil {
		t.Errorf("CreateMetricDescriptor(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("CreateMetricDescriptor(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_ListMonitoredResourceDescriptors_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListMonitoredResourceDescriptorsRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}}
	response, err := client.ListMonitoredResourceDescriptors(ctx, in)
	if err == nil {
		t.Errorf("ListMonitoredResourceDescriptors(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("ListMonitoredResourceDescriptors(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_ListTimeSeries_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.ListTimeSeriesRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}, "Filter": {}, "View": {}, "Interval": {}}
	response, err := client.ListTimeSeries(ctx, in)
	if err == nil {
		t.Errorf("ListTimeSeries(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("ListTimeSeries(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}

func TestMockMetricServer_CreateTimeSeries_MissingFieldsError(t *testing.T) {
	setup()
	defer tearDown()

	in := &monitoring.CreateTimeSeriesRequest{}
	want := validation.ErrMissingField.Err()
	missingFields := map[string]struct{}{"Name": {}, "TimeSeries": {}}
	response, err := client.CreateTimeSeries(ctx, in)
	if err == nil {
		t.Errorf("CreateTimeSeries(%q) == %q, expected error %q", in, response, want)
	}

	if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("CreateTimeSeries(%q) returned error %q, expected error %q",
			in, err.Error(), want)
	}

	if valid := validation.ValidateErrDetails(err, missingFields); !valid {
		t.Errorf("Expected missing fields %q", missingFields)
	}
}
