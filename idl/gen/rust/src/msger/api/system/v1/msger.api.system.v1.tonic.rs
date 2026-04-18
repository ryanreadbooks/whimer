// @generated
/// Generated client implementations.
pub mod chat_service_client {
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
    pub struct ChatServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ChatServiceClient<tonic::transport::Channel> {
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
    impl<T> ChatServiceClient<T>
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
        ) -> ChatServiceClient<InterceptedService<T, F>>
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
            ChatServiceClient::new(InterceptedService::new(inner, interceptor))
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
        pub async fn list_system_notify_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::ListSystemNotifyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
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
                "/msger.api.system.v1.ChatService/ListSystemNotifyMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.ChatService",
                        "ListSystemNotifyMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_system_reply_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::ListSystemReplyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
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
                "/msger.api.system.v1.ChatService/ListSystemReplyMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.ChatService",
                        "ListSystemReplyMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_system_mention_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::ListSystemMentionMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
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
                "/msger.api.system.v1.ChatService/ListSystemMentionMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.ChatService",
                        "ListSystemMentionMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn list_system_likes_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::ListSystemLikesMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
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
                "/msger.api.system.v1.ChatService/ListSystemLikesMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.ChatService",
                        "ListSystemLikesMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_chat_unread(
            &mut self,
            request: impl tonic::IntoRequest<super::GetChatUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetChatUnreadResponse>,
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
                "/msger.api.system.v1.ChatService/GetChatUnread",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("msger.api.system.v1.ChatService", "GetChatUnread"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_all_chats_unread(
            &mut self,
            request: impl tonic::IntoRequest<super::GetAllChatsUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetAllChatsUnreadResponse>,
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
                "/msger.api.system.v1.ChatService/GetAllChatsUnread",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.ChatService",
                        "GetAllChatsUnread",
                    ),
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
                "/msger.api.system.v1.ChatService/ClearChatUnread",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("msger.api.system.v1.ChatService", "ClearChatUnread"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn delete_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteMsgResponse>,
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
                "/msger.api.system.v1.ChatService/DeleteMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("msger.api.system.v1.ChatService", "DeleteMsg"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod chat_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with ChatServiceServer.
    #[async_trait]
    pub trait ChatService: std::marker::Send + std::marker::Sync + 'static {
        async fn list_system_notify_msg(
            &self,
            request: tonic::Request<super::ListSystemNotifyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
            tonic::Status,
        >;
        async fn list_system_reply_msg(
            &self,
            request: tonic::Request<super::ListSystemReplyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
            tonic::Status,
        >;
        async fn list_system_mention_msg(
            &self,
            request: tonic::Request<super::ListSystemMentionMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
            tonic::Status,
        >;
        async fn list_system_likes_msg(
            &self,
            request: tonic::Request<super::ListSystemLikesMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListSystemMsgResponse>,
            tonic::Status,
        >;
        async fn get_chat_unread(
            &self,
            request: tonic::Request<super::GetChatUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetChatUnreadResponse>,
            tonic::Status,
        >;
        async fn get_all_chats_unread(
            &self,
            request: tonic::Request<super::GetAllChatsUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetAllChatsUnreadResponse>,
            tonic::Status,
        >;
        async fn clear_chat_unread(
            &self,
            request: tonic::Request<super::ClearChatUnreadRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ClearChatUnreadResponse>,
            tonic::Status,
        >;
        async fn delete_msg(
            &self,
            request: tonic::Request<super::DeleteMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteMsgResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct ChatServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> ChatServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for ChatServiceServer<T>
    where
        T: ChatService,
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
                "/msger.api.system.v1.ChatService/ListSystemNotifyMsg" => {
                    #[allow(non_camel_case_types)]
                    struct ListSystemNotifyMsgSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::ListSystemNotifyMsgRequest>
                    for ListSystemNotifyMsgSvc<T> {
                        type Response = super::ListSystemMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListSystemNotifyMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::list_system_notify_msg(&inner, request)
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
                        let method = ListSystemNotifyMsgSvc(inner);
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
                "/msger.api.system.v1.ChatService/ListSystemReplyMsg" => {
                    #[allow(non_camel_case_types)]
                    struct ListSystemReplyMsgSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::ListSystemReplyMsgRequest>
                    for ListSystemReplyMsgSvc<T> {
                        type Response = super::ListSystemMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListSystemReplyMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::list_system_reply_msg(&inner, request)
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
                        let method = ListSystemReplyMsgSvc(inner);
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
                "/msger.api.system.v1.ChatService/ListSystemMentionMsg" => {
                    #[allow(non_camel_case_types)]
                    struct ListSystemMentionMsgSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::ListSystemMentionMsgRequest>
                    for ListSystemMentionMsgSvc<T> {
                        type Response = super::ListSystemMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListSystemMentionMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::list_system_mention_msg(&inner, request)
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
                        let method = ListSystemMentionMsgSvc(inner);
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
                "/msger.api.system.v1.ChatService/ListSystemLikesMsg" => {
                    #[allow(non_camel_case_types)]
                    struct ListSystemLikesMsgSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::ListSystemLikesMsgRequest>
                    for ListSystemLikesMsgSvc<T> {
                        type Response = super::ListSystemMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListSystemLikesMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::list_system_likes_msg(&inner, request)
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
                        let method = ListSystemLikesMsgSvc(inner);
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
                "/msger.api.system.v1.ChatService/GetChatUnread" => {
                    #[allow(non_camel_case_types)]
                    struct GetChatUnreadSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::GetChatUnreadRequest>
                    for GetChatUnreadSvc<T> {
                        type Response = super::GetChatUnreadResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetChatUnreadRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::get_chat_unread(&inner, request).await
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
                        let method = GetChatUnreadSvc(inner);
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
                "/msger.api.system.v1.ChatService/GetAllChatsUnread" => {
                    #[allow(non_camel_case_types)]
                    struct GetAllChatsUnreadSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::GetAllChatsUnreadRequest>
                    for GetAllChatsUnreadSvc<T> {
                        type Response = super::GetAllChatsUnreadResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetAllChatsUnreadRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::get_all_chats_unread(&inner, request)
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
                        let method = GetAllChatsUnreadSvc(inner);
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
                "/msger.api.system.v1.ChatService/ClearChatUnread" => {
                    #[allow(non_camel_case_types)]
                    struct ClearChatUnreadSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
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
                                <T as ChatService>::clear_chat_unread(&inner, request).await
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
                "/msger.api.system.v1.ChatService/DeleteMsg" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteMsgSvc<T: ChatService>(pub Arc<T>);
                    impl<
                        T: ChatService,
                    > tonic::server::UnaryService<super::DeleteMsgRequest>
                    for DeleteMsgSvc<T> {
                        type Response = super::DeleteMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as ChatService>::delete_msg(&inner, request).await
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
                        let method = DeleteMsgSvc(inner);
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
    impl<T> Clone for ChatServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "msger.api.system.v1.ChatService";
    impl<T> tonic::server::NamedService for ChatServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
/// Generated client implementations.
pub mod notification_service_client {
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
    pub struct NotificationServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl NotificationServiceClient<tonic::transport::Channel> {
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
    impl<T> NotificationServiceClient<T>
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
        ) -> NotificationServiceClient<InterceptedService<T, F>>
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
            NotificationServiceClient::new(InterceptedService::new(inner, interceptor))
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
        pub async fn notify_system_notice(
            &mut self,
            request: impl tonic::IntoRequest<super::NotifySystemNoticeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifySystemNoticeResponse>,
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
                "/msger.api.system.v1.NotificationService/NotifySystemNotice",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.NotificationService",
                        "NotifySystemNotice",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn notify_reply_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::NotifyReplyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyReplyMsgResponse>,
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
                "/msger.api.system.v1.NotificationService/NotifyReplyMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.NotificationService",
                        "NotifyReplyMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn notify_mention_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::NotifyMentionMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyMentionMsgResponse>,
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
                "/msger.api.system.v1.NotificationService/NotifyMentionMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.NotificationService",
                        "NotifyMentionMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn notify_likes_msg(
            &mut self,
            request: impl tonic::IntoRequest<super::NotifyLikesMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyLikesMsgResponse>,
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
                "/msger.api.system.v1.NotificationService/NotifyLikesMsg",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "msger.api.system.v1.NotificationService",
                        "NotifyLikesMsg",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod notification_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with NotificationServiceServer.
    #[async_trait]
    pub trait NotificationService: std::marker::Send + std::marker::Sync + 'static {
        async fn notify_system_notice(
            &self,
            request: tonic::Request<super::NotifySystemNoticeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifySystemNoticeResponse>,
            tonic::Status,
        >;
        async fn notify_reply_msg(
            &self,
            request: tonic::Request<super::NotifyReplyMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyReplyMsgResponse>,
            tonic::Status,
        >;
        async fn notify_mention_msg(
            &self,
            request: tonic::Request<super::NotifyMentionMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyMentionMsgResponse>,
            tonic::Status,
        >;
        async fn notify_likes_msg(
            &self,
            request: tonic::Request<super::NotifyLikesMsgRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NotifyLikesMsgResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct NotificationServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> NotificationServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for NotificationServiceServer<T>
    where
        T: NotificationService,
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
                "/msger.api.system.v1.NotificationService/NotifySystemNotice" => {
                    #[allow(non_camel_case_types)]
                    struct NotifySystemNoticeSvc<T: NotificationService>(pub Arc<T>);
                    impl<
                        T: NotificationService,
                    > tonic::server::UnaryService<super::NotifySystemNoticeRequest>
                    for NotifySystemNoticeSvc<T> {
                        type Response = super::NotifySystemNoticeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::NotifySystemNoticeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NotificationService>::notify_system_notice(
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
                        let method = NotifySystemNoticeSvc(inner);
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
                "/msger.api.system.v1.NotificationService/NotifyReplyMsg" => {
                    #[allow(non_camel_case_types)]
                    struct NotifyReplyMsgSvc<T: NotificationService>(pub Arc<T>);
                    impl<
                        T: NotificationService,
                    > tonic::server::UnaryService<super::NotifyReplyMsgRequest>
                    for NotifyReplyMsgSvc<T> {
                        type Response = super::NotifyReplyMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::NotifyReplyMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NotificationService>::notify_reply_msg(
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
                        let method = NotifyReplyMsgSvc(inner);
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
                "/msger.api.system.v1.NotificationService/NotifyMentionMsg" => {
                    #[allow(non_camel_case_types)]
                    struct NotifyMentionMsgSvc<T: NotificationService>(pub Arc<T>);
                    impl<
                        T: NotificationService,
                    > tonic::server::UnaryService<super::NotifyMentionMsgRequest>
                    for NotifyMentionMsgSvc<T> {
                        type Response = super::NotifyMentionMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::NotifyMentionMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NotificationService>::notify_mention_msg(
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
                        let method = NotifyMentionMsgSvc(inner);
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
                "/msger.api.system.v1.NotificationService/NotifyLikesMsg" => {
                    #[allow(non_camel_case_types)]
                    struct NotifyLikesMsgSvc<T: NotificationService>(pub Arc<T>);
                    impl<
                        T: NotificationService,
                    > tonic::server::UnaryService<super::NotifyLikesMsgRequest>
                    for NotifyLikesMsgSvc<T> {
                        type Response = super::NotifyLikesMsgResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::NotifyLikesMsgRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NotificationService>::notify_likes_msg(
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
                        let method = NotifyLikesMsgSvc(inner);
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
    impl<T> Clone for NotificationServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "msger.api.system.v1.NotificationService";
    impl<T> tonic::server::NamedService for NotificationServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
