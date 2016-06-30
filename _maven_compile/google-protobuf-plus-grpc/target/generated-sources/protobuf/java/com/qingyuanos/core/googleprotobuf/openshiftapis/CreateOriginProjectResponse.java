// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: manage_service.proto

package com.qingyuanos.core.googleprotobuf.openshiftapis;

/**
 * <pre>
 *google.protobuf.Any project = 1;
 * </pre>
 *
 * Protobuf type {@code openshift.CreateOriginProjectResponse}
 */
public  final class CreateOriginProjectResponse extends
    com.google.protobuf.GeneratedMessage implements
    // @@protoc_insertion_point(message_implements:openshift.CreateOriginProjectResponse)
    CreateOriginProjectResponseOrBuilder {
  // Use CreateOriginProjectResponse.newBuilder() to construct.
  private CreateOriginProjectResponse(com.google.protobuf.GeneratedMessage.Builder<?> builder) {
    super(builder);
  }
  private CreateOriginProjectResponse() {
    id_ = "";
    phase_ = "";
  }

  @java.lang.Override
  public final com.google.protobuf.UnknownFieldSet
  getUnknownFields() {
    return com.google.protobuf.UnknownFieldSet.getDefaultInstance();
  }
  private CreateOriginProjectResponse(
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
            java.lang.String s = input.readStringRequireUtf8();

            id_ = s;
            break;
          }
          case 18: {
            java.lang.String s = input.readStringRequireUtf8();

            phase_ = s;
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
    return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectResponse_descriptor;
  }

  protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectResponse_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.class, com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.Builder.class);
  }

  public static final int ID_FIELD_NUMBER = 1;
  private volatile java.lang.Object id_;
  /**
   * <code>optional string id = 1;</code>
   */
  public java.lang.String getId() {
    java.lang.Object ref = id_;
    if (ref instanceof java.lang.String) {
      return (java.lang.String) ref;
    } else {
      com.google.protobuf.ByteString bs = 
          (com.google.protobuf.ByteString) ref;
      java.lang.String s = bs.toStringUtf8();
      id_ = s;
      return s;
    }
  }
  /**
   * <code>optional string id = 1;</code>
   */
  public com.google.protobuf.ByteString
      getIdBytes() {
    java.lang.Object ref = id_;
    if (ref instanceof java.lang.String) {
      com.google.protobuf.ByteString b = 
          com.google.protobuf.ByteString.copyFromUtf8(
              (java.lang.String) ref);
      id_ = b;
      return b;
    } else {
      return (com.google.protobuf.ByteString) ref;
    }
  }

  public static final int PHASE_FIELD_NUMBER = 2;
  private volatile java.lang.Object phase_;
  /**
   * <code>optional string phase = 2;</code>
   */
  public java.lang.String getPhase() {
    java.lang.Object ref = phase_;
    if (ref instanceof java.lang.String) {
      return (java.lang.String) ref;
    } else {
      com.google.protobuf.ByteString bs = 
          (com.google.protobuf.ByteString) ref;
      java.lang.String s = bs.toStringUtf8();
      phase_ = s;
      return s;
    }
  }
  /**
   * <code>optional string phase = 2;</code>
   */
  public com.google.protobuf.ByteString
      getPhaseBytes() {
    java.lang.Object ref = phase_;
    if (ref instanceof java.lang.String) {
      com.google.protobuf.ByteString b = 
          com.google.protobuf.ByteString.copyFromUtf8(
              (java.lang.String) ref);
      phase_ = b;
      return b;
    } else {
      return (com.google.protobuf.ByteString) ref;
    }
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
    if (!getIdBytes().isEmpty()) {
      com.google.protobuf.GeneratedMessage.writeString(output, 1, id_);
    }
    if (!getPhaseBytes().isEmpty()) {
      com.google.protobuf.GeneratedMessage.writeString(output, 2, phase_);
    }
  }

  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    if (!getIdBytes().isEmpty()) {
      size += com.google.protobuf.GeneratedMessage.computeStringSize(1, id_);
    }
    if (!getPhaseBytes().isEmpty()) {
      size += com.google.protobuf.GeneratedMessage.computeStringSize(2, phase_);
    }
    memoizedSize = size;
    return size;
  }

  private static final long serialVersionUID = 0L;
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parseFrom(
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
  public static Builder newBuilder(com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse prototype) {
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
   * Protobuf type {@code openshift.CreateOriginProjectResponse}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessage.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:openshift.CreateOriginProjectResponse)
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponseOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectResponse_descriptor;
    }

    protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectResponse_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.class, com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.Builder.class);
    }

    // Construct using com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.newBuilder()
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
      id_ = "";

      phase_ = "";

      return this;
    }

    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.ProjectAndBuild.internal_static_openshift_CreateOriginProjectResponse_descriptor;
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse getDefaultInstanceForType() {
      return com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.getDefaultInstance();
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse build() {
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse buildPartial() {
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse result = new com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse(this);
      result.id_ = id_;
      result.phase_ = phase_;
      onBuilt();
      return result;
    }

    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse) {
        return mergeFrom((com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse other) {
      if (other == com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse.getDefaultInstance()) return this;
      if (!other.getId().isEmpty()) {
        id_ = other.id_;
        onChanged();
      }
      if (!other.getPhase().isEmpty()) {
        phase_ = other.phase_;
        onChanged();
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
      com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse parsedMessage = null;
      try {
        parsedMessage = PARSER.parsePartialFrom(input, extensionRegistry);
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        parsedMessage = (com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse) e.getUnfinishedMessage();
        throw e.unwrapIOException();
      } finally {
        if (parsedMessage != null) {
          mergeFrom(parsedMessage);
        }
      }
      return this;
    }

    private java.lang.Object id_ = "";
    /**
     * <code>optional string id = 1;</code>
     */
    public java.lang.String getId() {
      java.lang.Object ref = id_;
      if (!(ref instanceof java.lang.String)) {
        com.google.protobuf.ByteString bs =
            (com.google.protobuf.ByteString) ref;
        java.lang.String s = bs.toStringUtf8();
        id_ = s;
        return s;
      } else {
        return (java.lang.String) ref;
      }
    }
    /**
     * <code>optional string id = 1;</code>
     */
    public com.google.protobuf.ByteString
        getIdBytes() {
      java.lang.Object ref = id_;
      if (ref instanceof String) {
        com.google.protobuf.ByteString b = 
            com.google.protobuf.ByteString.copyFromUtf8(
                (java.lang.String) ref);
        id_ = b;
        return b;
      } else {
        return (com.google.protobuf.ByteString) ref;
      }
    }
    /**
     * <code>optional string id = 1;</code>
     */
    public Builder setId(
        java.lang.String value) {
      if (value == null) {
    throw new NullPointerException();
  }
  
      id_ = value;
      onChanged();
      return this;
    }
    /**
     * <code>optional string id = 1;</code>
     */
    public Builder clearId() {
      
      id_ = getDefaultInstance().getId();
      onChanged();
      return this;
    }
    /**
     * <code>optional string id = 1;</code>
     */
    public Builder setIdBytes(
        com.google.protobuf.ByteString value) {
      if (value == null) {
    throw new NullPointerException();
  }
  checkByteStringIsUtf8(value);
      
      id_ = value;
      onChanged();
      return this;
    }

    private java.lang.Object phase_ = "";
    /**
     * <code>optional string phase = 2;</code>
     */
    public java.lang.String getPhase() {
      java.lang.Object ref = phase_;
      if (!(ref instanceof java.lang.String)) {
        com.google.protobuf.ByteString bs =
            (com.google.protobuf.ByteString) ref;
        java.lang.String s = bs.toStringUtf8();
        phase_ = s;
        return s;
      } else {
        return (java.lang.String) ref;
      }
    }
    /**
     * <code>optional string phase = 2;</code>
     */
    public com.google.protobuf.ByteString
        getPhaseBytes() {
      java.lang.Object ref = phase_;
      if (ref instanceof String) {
        com.google.protobuf.ByteString b = 
            com.google.protobuf.ByteString.copyFromUtf8(
                (java.lang.String) ref);
        phase_ = b;
        return b;
      } else {
        return (com.google.protobuf.ByteString) ref;
      }
    }
    /**
     * <code>optional string phase = 2;</code>
     */
    public Builder setPhase(
        java.lang.String value) {
      if (value == null) {
    throw new NullPointerException();
  }
  
      phase_ = value;
      onChanged();
      return this;
    }
    /**
     * <code>optional string phase = 2;</code>
     */
    public Builder clearPhase() {
      
      phase_ = getDefaultInstance().getPhase();
      onChanged();
      return this;
    }
    /**
     * <code>optional string phase = 2;</code>
     */
    public Builder setPhaseBytes(
        com.google.protobuf.ByteString value) {
      if (value == null) {
    throw new NullPointerException();
  }
  checkByteStringIsUtf8(value);
      
      phase_ = value;
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


    // @@protoc_insertion_point(builder_scope:openshift.CreateOriginProjectResponse)
  }

  // @@protoc_insertion_point(class_scope:openshift.CreateOriginProjectResponse)
  private static final com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse();
  }

  public static com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<CreateOriginProjectResponse>
      PARSER = new com.google.protobuf.AbstractParser<CreateOriginProjectResponse>() {
    public CreateOriginProjectResponse parsePartialFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws com.google.protobuf.InvalidProtocolBufferException {
        return new CreateOriginProjectResponse(input, extensionRegistry);
    }
  };

  public static com.google.protobuf.Parser<CreateOriginProjectResponse> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<CreateOriginProjectResponse> getParserForType() {
    return PARSER;
  }

  public com.qingyuanos.core.googleprotobuf.openshiftapis.CreateOriginProjectResponse getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

