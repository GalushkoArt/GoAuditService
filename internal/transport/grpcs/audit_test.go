package grpcs

import (
	"context"
	"errors"
	"github.com/galushkoart/go-audit-service/mock"
	"github.com/galushkoart/go-audit-service/pkg/model"
	audit "github.com/galushkoart/go-audit-service/pkg/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

//go:generate echo $PWD - $GOFILE
//go:generate mockgen -package mock -destination ../../../mock/audit_service_mock.go -source=../../service/audit.go AuditService

func TestAuditHandler(t *testing.T) {
	t.Parallel()
	mockService := mock.NewMockAuditService(gomock.NewController(t))
	handler := newAuditHandler(mockService)
	for _, td := range auditHandlerTestData {
		t.Run(td.name, func(t *testing.T) {
			mockService.EXPECT().Insert(gomock.Any(), model.LogRequestToItem(td.request)).Return(td.errorFromService)
			resp, err := handler.Log(context.Background(), td.request)
			assert.Equal(t, td.expectedError, err, "Error should be equal")
			assert.Equal(t, td.response, resp, "Response should be equal")
		})
	}
}

var auditHandlerTestData = []struct {
	name             string
	request          *audit.LogRequest
	errorFromService error
	response         *audit.Response
	expectedError    error
}{
	{
		name:             "Success",
		request:          &audit.LogRequest{Action: audit.LogRequest_SIGN_IN},
		errorFromService: nil,
		response:         &audit.Response{Answer: audit.Response_SUCCESS},
		expectedError:    nil,
	},
	{
		name:             "Error",
		request:          &audit.LogRequest{Action: audit.LogRequest_SIGN_IN},
		errorFromService: errors.New("failed to insert log"),
		response:         &audit.Response{Answer: audit.Response_ERROR},
		expectedError:    status.Error(codes.Internal, "failed to insert log"),
	},
}
