// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: manage_service.proto

package com.qingyuanos.core.googleprotobuf.openshiftapis;

/**
 * <pre>
 *google.protobuf.Any project = 1;
 * </pre>
 *
 * Protobuf type {@code openshift.CreateOriginProjectArbitraryRequest}
 */
public  final class CreateOriginProjectArbitraryRequest extends
    com.google.protobuf.GeneratedMessage implements
    // @@protoc_insertion_point(message_implements:openshift.CreateOriginProjectArbitraryRequest)
    CreateOriginProjectArbitraryRequestOrBuilder {
  // Use CreateOriginProjectArbitraryRequest.newBuilder() to construct.
  private CreateOriginProjectArbitraryRequest(com.google.protobuf.GeneratedMessage.Builder<?> builder) {
    super(builder);
  }
  private CreateOriginProjectArbitraryRequest() {
    odefv1RawData_ = com.google.protobuf.ByteString.EMPTY;
  }

  @java.lang.Override
  public final com.google.protobuf.UnknownFieldSet
  getUnknownFields() {
    return com.google.protobuf.UnknownFieldSet.getDefaultInstance();
  }
  private CreateOriginProjectArbitraryRequest(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    this();
    int mutable_bitField0_ = 0;
    try {
      boolean done = false;
      while (!done) {
        int tag = input.readTag();
        switch (tag) {
          case 0:
            done = true;
            break;
          default: {
            if (!input.skipField(tag)) {
              done = true;
            }
            break;
          }
          case 10: {

            odefv1RawData_ = input.readBytes();
            break;
          }
        }
      }
    } catch (com.google.protobuf.InvalidProtocolBufferException e) {
      throw e.setUnfinishedMessage(this);
    } catch (java.io.IOException e) {
      throw new com.google.protobuf.InvalidProtocolBufferException(
          e).setUnfinishedMessage(this);
    } finally {
      makeExtensionsImmutable();
    }
  }
  public static final com.google.protobuf.Descriptors.Descriptor
      getDescriptor() {
    return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectArbitraryRequest_descriptor;
  }

  protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectArbitraryRequest_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.class, com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.Builder.class);
  }

  public static final int ODEFV1RAWDATA_FIELD_NUMBER = 1;
  private com.google.protobuf.ByteString odefv1RawData_;
  /**
   * <code>optional bytes odefv1RawData = 1;</code>
   */
  public com.google.protobuf.ByteString getOdefv1RawData() {
    return odefv1RawData_;
  }

  private byte memoizedIsInitialized = -1;
  public final boolean isInitialized() {
    byte isInitialized = memoizedIsInitialized;
    if (isInitialized == 1) return true;
    if (isInitialized == 0) return false;

    memoizedIsInitialized = 1;
    return true;
  }

  public void writeTo(com.google.protobuf.CodedOutputStream output)
                      throws java.io.IOException {
    if (!odefv1RawData_.isEmpty()) {
      output.writeBytes(1, odefv1RawData_);
    }
  }

  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    if (!odefv1RawData_.isEmpty()) {
      size += com.google.protobuf.CodedOutputStream
        .computeBytesSize(1, odefv1RawData_);
    }
    memoizedSize = size;
    return size;
  }

  private static final long serialVersionUID = 0L;
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parseFrom(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  public Builder newBuilderForType() { return newBuilder(); }
  public static Builder newBuilder() {
    return DEFAULT_INSTANCE.toBuilder();
  }
  public static Builder newBuilder(com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest prototype) {
    return DEFAULT_INSTANCE.toBuilder().mergeFrom(prototype);
  }
  public Builder toBuilder() {
    return this == DEFAULT_INSTANCE
        ? new Builder() : new Builder().mergeFrom(this);
  }

  @java.lang.Override
  protected Builder newBuilderForType(
      com.google.protobuf.GeneratedMessage.BuilderParent parent) {
    Builder builder = new Builder(parent);
    return builder;
  }
  /**
   * <pre>
   *google.protobuf.Any project = 1;
   * </pre>
   *
   * Protobuf type {@code openshift.CreateOriginProjectArbitraryRequest}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessage.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:openshift.CreateOriginProjectArbitraryRequest)
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequestOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectArbitraryRequest_descriptor;
    }

    protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectArbitraryRequest_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.class, com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.Builder.class);
    }

    // Construct using com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.newBuilder()
    private Builder() {
      maybeForceBuilderInitialization();
    }

    private Builder(
        com.google.protobuf.GeneratedMessage.BuilderParent parent) {
      super(parent);
      maybeForceBuilderInitialization();
    }
    private void maybeForceBuilderInitialization() {
      if (com.google.protobuf.GeneratedMessage.alwaysUseFieldBuilders) {
      }
    }
    public Builder clear() {
      super.clear();
      odefv1RawData_ = com.google.protobuf.ByteString.EMPTY;

      return this;
    }

    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectArbitraryRequest_descriptor;
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest getDefaultInstanceForType() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.getDefaultInstance();
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest build() {
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest buildPartial() {
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest result = new com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest(this);
      result.odefv1RawData_ = odefv1RawData_;
      onBuilt();
      return result;
    }

    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest) {
        return mergeFrom((com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest other) {
      if (other == com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest.getDefaultInstance()) return this;
      if (other.getOdefv1RawData() != com.google.protobuf.ByteString.EMPTY) {
        setOdefv1RawData(other.getOdefv1RawData());
      }
      onChanged();
      return this;
    }

    public final boolean isInitialized() {
      return true;
    }

    public Builder mergeFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws java.io.IOException {
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest parsedMessage = null;
      try {
        parsedMessage = PARSER.parsePartialFrom(input, extensionRegistry);
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        parsedMessage = (com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest) e.getUnfinishedMessage();
        throw e.unwrapIOException();
      } finally {
        if (parsedMessage != null) {
          mergeFrom(parsedMessage);
        }
      }
      return this;
    }

    private com.google.protobuf.ByteString odefv1RawData_ = com.google.protobuf.ByteString.EMPTY;
    /**
     * <code>optional bytes odefv1RawData = 1;</code>
     */
    public com.google.protobuf.ByteString getOdefv1RawData() {
      return odefv1RawData_;
    }
    /**
     * <code>optional bytes odefv1RawData = 1;</code>
     */
    public Builder setOdefv1RawData(com.google.protobuf.ByteString value) {
      if (value == null) {
    throw new NullPointerException();
  }
  
      odefv1RawData_ = value;
      onChanged();
      return this;
    }
    /**
     * <code>optional bytes odefv1RawData = 1;</code>
     */
    public Builder clearOdefv1RawData() {
      
      odefv1RawData_ = getDefaultInstance().getOdefv1RawData();
      onChanged();
      return this;
    }
    public final Builder setUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return this;
    }

    public final Builder mergeUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return this;
    }


    // @@protoc_insertion_point(builder_scope:openshift.CreateOriginProjectArbitraryRequest)
  }

  // @@protoc_insertion_point(class_scope:openshift.CreateOriginProjectArbitraryRequest)
  private static final com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest();
  }

  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<CreateOriginProjectArbitraryRequest>
      PARSER = new com.google.protobuf.AbstractParser<CreateOriginProjectArbitraryRequest>() {
    public CreateOriginProjectArbitraryRequest parsePartialFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws com.google.protobuf.InvalidProtocolBufferException {
        return new CreateOriginProjectArbitraryRequest(input, extensionRegistry);
    }
  };

  public static com.google.protobuf.Parser<CreateOriginProjectArbitraryRequest> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<CreateOriginProjectArbitraryRequest> getParserForType() {
    return PARSER;
  }

  public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectArbitraryRequest getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

