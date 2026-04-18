// @generated
/// Generated client implementations.
pub mod user_chat_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct UserChatServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl UserChatServiceClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> UserChatServiceClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::Body>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + std::marker::Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + std::marker::Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> UserChatServiceClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::Body>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::Body>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::Body>,
            >>::Error: Into<StdError> + std::marker::Send + std::marker::Sync,
        {
            UserChatServiceClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        pub async fn create_p2p_chat(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateP2pChatRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateP2pChatResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/CreateP2PChat",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "CreateP2PChat",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn send_msg_to_chat(
            &mut self,
            request: impl tonic::IntoRequest<super::SendMsgToChatRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SendMsgToChatResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/SendMsgToChat",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "SendMsgToChat",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_chat_members(
            &mut self,
            request: impl tonic::IntoRequest<super::GetChatMembersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetChatMembersResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/GetChatMembers",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "GetChatMembers",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn batch_get_chat_members(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchGetChatMembersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetChatMembersResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/BatchGetChatMembers",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "BatchGetChatMembers",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_recent_chats(
            &mut self,
            request: impl tonic::IntoRequest<super::ListRecentChatsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListRecentChatsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/ListRecentChats",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "ListRecentChats",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_chat_msgs(
            &mut self,
            request: impl tonic::IntoRequest<super::ListChatMsgsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListChatMsgsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/ListChatMsgs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "ListChatMsgs",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn recall_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::RecallMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RecallMsgResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/RecallMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("msger.api.userchat.v1.UserChatService", "RecallMsg"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn clear_chat_unread(
            &mut self,
            request: impl tonic::IntoRequest<super::ClearChatUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ClearChatUnreadResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/msger.api.userchat.v1.UserChatService/ClearChatUnread",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.userchat.v1.UserChatService",
                        "ClearChatUnread",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod user_chat_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with UserChatServiceServer.
    #[async_trait]
    pub trait UserChatService: std::marker::Send + std::marker::Sync + 'static {
        async fn create_p2p_chat(
            &self,
            request: tonic::Request<super::CreateP2pChatRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateP2pChatResponse>,
            tonic::Status,
        >;
        async fn send_msg_to_chat(
            &self,
            request: tonic::Request<super::SendMsgToChatRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SendMsgToChatResponse>,
            tonic::Status,
        >;
        async fn get_chat_members(
            &self,
            request: tonic::Request<super::GetChatMembersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetChatMembersResponse>,
            tonic::Status,
        >;
        async fn batch_get_chat_members(
            &self,
            request: tonic::Request<super::BatchGetChatMembersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetChatMembersResponse>,
            tonic::Status,
        >;
        async fn list_recent_chats(
            &self,
            request: tonic::Request<super::ListRecentChatsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListRecentChatsResponse>,
            tonic::Status,
        >;
        async fn list_chat_msgs(
            &self,
            request: tonic::Request<super::ListChatMsgsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListChatMsgsResponse>,
            tonic::Status,
        >;
        async fn recall_msg(
            &self,
            request: tonic::Request<super::RecallMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RecallMsgResponse>,
            tonic::Status,
        >;
        async fn clear_chat_unread(
            &self,
            request: tonic::Request<super::ClearChatUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ClearChatUnreadResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct UserChatServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> UserChatServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for UserChatServiceServer<T>
    where
        T: UserChatService,
        B: Body + std::marker::Send + 'static,
        B::Error: Into<StdError> + std::marker::Send + 'static,
    {
        type Response = http::Response<tonic::body::Body>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            match req.uri().path() {
                "/msger.api.userchat.v1.UserChatService/CreateP2PChat" => {
                    #[allow(non_camel_case_types)]
                    struct CreateP2PChatSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::CreateP2pChatRequest>
                    for CreateP2PChatSvc<T> {
                        type Response = super::CreateP2pChatResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateP2pChatRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::create_p2p_chat(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreateP2PChatSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/SendMsgToChat" => {
                    #[allow(non_camel_case_types)]
                    struct SendMsgToChatSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::SendMsgToChatRequest>
                    for SendMsgToChatSvc<T> {
                        type Response = super::SendMsgToChatResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::SendMsgToChatRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::send_msg_to_chat(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = SendMsgToChatSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/GetChatMembers" => {
                    #[allow(non_camel_case_types)]
                    struct GetChatMembersSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::GetChatMembersRequest>
                    for GetChatMembersSvc<T> {
                        type Response = super::GetChatMembersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetChatMembersRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::get_chat_members(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = GetChatMembersSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/BatchGetChatMembers" => {
                    #[allow(non_camel_case_types)]
                    struct BatchGetChatMembersSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::BatchGetChatMembersRequest>
                    for BatchGetChatMembersSvc<T> {
                        type Response = super::BatchGetChatMembersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchGetChatMembersRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::batch_get_chat_members(
                                        &inner,
                                        request,
                                    )
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = BatchGetChatMembersSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/ListRecentChats" => {
                    #[allow(non_camel_case_types)]
                    struct ListRecentChatsSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::ListRecentChatsRequest>
                    for ListRecentChatsSvc<T> {
                        type Response = super::ListRecentChatsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListRecentChatsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::list_recent_chats(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListRecentChatsSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/ListChatMsgs" => {
                    #[allow(non_camel_case_types)]
                    struct ListChatMsgsSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::ListChatMsgsRequest>
                    for ListChatMsgsSvc<T> {
                        type Response = super::ListChatMsgsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListChatMsgsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::list_chat_msgs(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListChatMsgsSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/RecallMsg" => {
                    #[allow(non_camel_case_types)]
                    struct RecallMsgSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::RecallMsgRequest>
                    for RecallMsgSvc<T> {
                        type Response = super::RecallMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RecallMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::recall_msg(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RecallMsgSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/msger.api.userchat.v1.UserChatService/ClearChatUnread" => {
                    #[allow(non_camel_case_types)]
                    struct ClearChatUnreadSvc<T: UserChatService>(pub Arc<T>);
                    impl<
                        T: UserChatService,
                    > tonic::server::UnaryService<super::ClearChatUnreadRequest>
                    for ClearChatUnreadSvc<T> {
                        type Response = super::ClearChatUnreadResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ClearChatUnreadRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as UserChatService>::clear_chat_unread(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ClearChatUnreadSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        let mut response = http::Response::new(
                            tonic::body::Body::default(),
                        );
                        let headers = response.headers_mut();
                        headers
                            .insert(
                                tonic::Status::GRPC_STATUS,
                                (tonic::Code::Unimplemented as i32).into(),
                            );
                        headers
                            .insert(
                                http::header::CONTENT_TYPE,
                                tonic::metadata::GRPC_CONTENT_TYPE,
                            );
                        Ok(response)
                    })
                }
            }
        }
    }
    impl<T> Clone for UserChatServiceServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    /// Generated gRPC service name
    pub const SERVICE_NAME: &str = "msger.api.userchat.v1.UserChatService";
    impl<T> tonic::server::NamedService for UserChatServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
