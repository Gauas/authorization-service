package schema

import (
	"sync"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	ServiceName             = "authorization.v1.AuthorizationService"
	ValidateTokenMethodName = "ValidateToken"
)

var (
	tokenSchemaOnce sync.Once

	validateTokenRequestDesc  protoreflect.MessageDescriptor
	validateTokenResponseDesc protoreflect.MessageDescriptor

	requestTokenField protoreflect.FieldDescriptor

	responseUserIDField     protoreflect.FieldDescriptor
	responsePermissionField protoreflect.FieldDescriptor
	responseDeviceIDField   protoreflect.FieldDescriptor
)

func NewValidateTokenRequest() *dynamicpb.Message {
	initTokenSchema()
	return dynamicpb.NewMessage(validateTokenRequestDesc)
}

func ValidateTokenRequestToken(message *dynamicpb.Message) string {
	initTokenSchema()
	return message.Get(requestTokenField).String()
}

func NewValidateTokenResponse(userID, permission, deviceID string) *dynamicpb.Message {
	initTokenSchema()

	message := dynamicpb.NewMessage(validateTokenResponseDesc)
	message.Set(responseUserIDField, protoreflect.ValueOfString(userID))
	message.Set(responsePermissionField, protoreflect.ValueOfString(permission))
	message.Set(responseDeviceIDField, protoreflect.ValueOfString(deviceID))

	return message
}

func initTokenSchema() {
	tokenSchemaOnce.Do(func() {
		fileDescriptorProto := &descriptorpb.FileDescriptorProto{
			Syntax:  stringPtr("proto3"),
			Name:    stringPtr("authorization/v1/authorization.proto"),
			Package: stringPtr("authorization.v1"),
			MessageType: []*descriptorpb.DescriptorProto{
				{
					Name: stringPtr("ValidateTokenRequest"),
					Field: []*descriptorpb.FieldDescriptorProto{
						stringField("token", 1),
					},
				},
				{
					Name: stringPtr("ValidateTokenResponse"),
					Field: []*descriptorpb.FieldDescriptorProto{
						stringField("user_id", 1),
						stringField("permission", 2),
						stringField("device_id", 3),
					},
				},
			},
			Service: []*descriptorpb.ServiceDescriptorProto{
				{
					Name: stringPtr("AuthorizationService"),
					Method: []*descriptorpb.MethodDescriptorProto{
						{
							Name:       stringPtr(ValidateTokenMethodName),
							InputType:  stringPtr(".authorization.v1.ValidateTokenRequest"),
							OutputType: stringPtr(".authorization.v1.ValidateTokenResponse"),
						},
					},
				},
			},
		}

		fileDescriptor, err := protodesc.NewFile(fileDescriptorProto, protoregistry.GlobalFiles)
		if err != nil {
			panic(err)
		}

		if err := protoregistry.GlobalFiles.RegisterFile(fileDescriptor); err != nil {
			panic(err)
		}

		validateTokenRequestDesc = fileDescriptor.Messages().ByName("ValidateTokenRequest")
		validateTokenResponseDesc = fileDescriptor.Messages().ByName("ValidateTokenResponse")

		requestTokenField = validateTokenRequestDesc.Fields().ByName("token")
		responseUserIDField = validateTokenResponseDesc.Fields().ByName("user_id")
		responsePermissionField = validateTokenResponseDesc.Fields().ByName("permission")
		responseDeviceIDField = validateTokenResponseDesc.Fields().ByName("device_id")
	})
}

func stringField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   stringPtr(name),
		Number: int32Ptr(number),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
	}
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
