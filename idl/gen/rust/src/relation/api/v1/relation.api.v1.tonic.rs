// @generated
/// Generated client implementations.
pub mod relation_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    ///
    #[derive(Debug, Clone)]
    pub struct RelationServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl RelationServiceClient<tonic::transport::Channel> {
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
    impl<T> RelationServiceClient<T>
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
        ) -> RelationServiceClient<InterceptedService<T, F>>
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
            RelationServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** 关注/取消关注某个用户
*/
        pub async fn follow_user(
            &mut self,
            request: impl tonic::IntoRequest<super::FollowUserRequest>,
        ) -> std::result::Result<
            tonic::Response<super::FollowUserResponse>,
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
                "/relation.api.v1.RelationService/FollowUser",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("relation.api.v1.RelationService", "FollowUser"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某个用户的粉丝列表
*/
        pub async fn get_user_fan_list(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserFanListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFanListResponse>,
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
                "/relation.api.v1.RelationService/GetUserFanList",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("relation.api.v1.RelationService", "GetUserFanList"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某个用户的关注列表
*/
        pub async fn get_user_following_list(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserFollowingListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFollowingListResponse>,
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
                "/relation.api.v1.RelationService/GetUserFollowingList",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "GetUserFollowingList",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 移除某个用户的粉丝
*/
        pub async fn remove_user_fan(
            &mut self,
            request: impl tonic::IntoRequest<super::RemoveUserFanRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RemoveUserFanResponse>,
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
                "/relation.api.v1.RelationService/RemoveUserFan",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("relation.api.v1.RelationService", "RemoveUserFan"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取用户的粉丝数量
*/
        pub async fn get_user_fan_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserFanCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFanCountResponse>,
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
                "/relation.api.v1.RelationService/GetUserFanCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("relation.api.v1.RelationService", "GetUserFanCount"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取用户的关注数量
*/
        pub async fn get_user_following_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserFollowingCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFollowingCountResponse>,
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
                "/relation.api.v1.RelationService/GetUserFollowingCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "GetUserFollowingCount",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 判断某个用户是否关注了某些用户
*/
        pub async fn batch_check_user_followed(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckUserFollowedRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserFollowedResponse>,
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
                "/relation.api.v1.RelationService/BatchCheckUserFollowed",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "BatchCheckUserFollowed",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn check_user_followed(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckUserFollowedRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserFollowedResponse>,
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
                "/relation.api.v1.RelationService/CheckUserFollowed",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "CheckUserFollowed",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 分页获取某个用户的粉丝列表
*/
        pub async fn page_get_user_fan_list(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetUserFanListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserFanListResponse>,
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
                "/relation.api.v1.RelationService/PageGetUserFanList",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "PageGetUserFanList",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 分页获取某个用户的关注列表
*/
        pub async fn page_get_user_following_list(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetUserFollowingListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserFollowingListResponse>,
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
                "/relation.api.v1.RelationService/PageGetUserFollowingList",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "PageGetUserFollowingList",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 关注设置
*/
        pub async fn update_user_settings(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateUserSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateUserSettingsResponse>,
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
                "/relation.api.v1.RelationService/UpdateUserSettings",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "relation.api.v1.RelationService",
                        "UpdateUserSettings",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_user_settings(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserSettingsResponse>,
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
                "/relation.api.v1.RelationService/GetUserSettings",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("relation.api.v1.RelationService", "GetUserSettings"),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod relation_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with RelationServiceServer.
    #[async_trait]
    pub trait RelationService: std::marker::Send + std::marker::Sync + 'static {
        /** 关注/取消关注某个用户
*/
        async fn follow_user(
            &self,
            request: tonic::Request<super::FollowUserRequest>,
        ) -> std::result::Result<
            tonic::Response<super::FollowUserResponse>,
            tonic::Status,
        >;
        /** 获取某个用户的粉丝列表
*/
        async fn get_user_fan_list(
            &self,
            request: tonic::Request<super::GetUserFanListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFanListResponse>,
            tonic::Status,
        >;
        /** 获取某个用户的关注列表
*/
        async fn get_user_following_list(
            &self,
            request: tonic::Request<super::GetUserFollowingListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFollowingListResponse>,
            tonic::Status,
        >;
        /** 移除某个用户的粉丝
*/
        async fn remove_user_fan(
            &self,
            request: tonic::Request<super::RemoveUserFanRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RemoveUserFanResponse>,
            tonic::Status,
        >;
        /** 获取用户的粉丝数量
*/
        async fn get_user_fan_count(
            &self,
            request: tonic::Request<super::GetUserFanCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFanCountResponse>,
            tonic::Status,
        >;
        /** 获取用户的关注数量
*/
        async fn get_user_following_count(
            &self,
            request: tonic::Request<super::GetUserFollowingCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserFollowingCountResponse>,
            tonic::Status,
        >;
        /** 判断某个用户是否关注了某些用户
*/
        async fn batch_check_user_followed(
            &self,
            request: tonic::Request<super::BatchCheckUserFollowedRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserFollowedResponse>,
            tonic::Status,
        >;
        ///
        async fn check_user_followed(
            &self,
            request: tonic::Request<super::CheckUserFollowedRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserFollowedResponse>,
            tonic::Status,
        >;
        /** 分页获取某个用户的粉丝列表
*/
        async fn page_get_user_fan_list(
            &self,
            request: tonic::Request<super::PageGetUserFanListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserFanListResponse>,
            tonic::Status,
        >;
        /** 分页获取某个用户的关注列表
*/
        async fn page_get_user_following_list(
            &self,
            request: tonic::Request<super::PageGetUserFollowingListRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserFollowingListResponse>,
            tonic::Status,
        >;
        /** 关注设置
*/
        async fn update_user_settings(
            &self,
            request: tonic::Request<super::UpdateUserSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateUserSettingsResponse>,
            tonic::Status,
        >;
        ///
        async fn get_user_settings(
            &self,
            request: tonic::Request<super::GetUserSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserSettingsResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct RelationServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> RelationServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for RelationServiceServer<T>
    where
        T: RelationService,
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
                "/relation.api.v1.RelationService/FollowUser" => {
                    #[allow(non_camel_case_types)]
                    struct FollowUserSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::FollowUserRequest>
                    for FollowUserSvc<T> {
                        type Response = super::FollowUserResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::FollowUserRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::follow_user(&inner, request).await
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
                        let method = FollowUserSvc(inner);
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
                "/relation.api.v1.RelationService/GetUserFanList" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserFanListSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::GetUserFanListRequest>
                    for GetUserFanListSvc<T> {
                        type Response = super::GetUserFanListResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserFanListRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::get_user_fan_list(&inner, request)
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
                        let method = GetUserFanListSvc(inner);
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
                "/relation.api.v1.RelationService/GetUserFollowingList" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserFollowingListSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::GetUserFollowingListRequest>
                    for GetUserFollowingListSvc<T> {
                        type Response = super::GetUserFollowingListResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserFollowingListRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::get_user_following_list(
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
                        let method = GetUserFollowingListSvc(inner);
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
                "/relation.api.v1.RelationService/RemoveUserFan" => {
                    #[allow(non_camel_case_types)]
                    struct RemoveUserFanSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::RemoveUserFanRequest>
                    for RemoveUserFanSvc<T> {
                        type Response = super::RemoveUserFanResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RemoveUserFanRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::remove_user_fan(&inner, request)
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
                        let method = RemoveUserFanSvc(inner);
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
                "/relation.api.v1.RelationService/GetUserFanCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserFanCountSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::GetUserFanCountRequest>
                    for GetUserFanCountSvc<T> {
                        type Response = super::GetUserFanCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserFanCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::get_user_fan_count(&inner, request)
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
                        let method = GetUserFanCountSvc(inner);
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
                "/relation.api.v1.RelationService/GetUserFollowingCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserFollowingCountSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::GetUserFollowingCountRequest>
                    for GetUserFollowingCountSvc<T> {
                        type Response = super::GetUserFollowingCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserFollowingCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::get_user_following_count(
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
                        let method = GetUserFollowingCountSvc(inner);
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
                "/relation.api.v1.RelationService/BatchCheckUserFollowed" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckUserFollowedSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::BatchCheckUserFollowedRequest>
                    for BatchCheckUserFollowedSvc<T> {
                        type Response = super::BatchCheckUserFollowedResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchCheckUserFollowedRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::batch_check_user_followed(
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
                        let method = BatchCheckUserFollowedSvc(inner);
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
                "/relation.api.v1.RelationService/CheckUserFollowed" => {
                    #[allow(non_camel_case_types)]
                    struct CheckUserFollowedSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::CheckUserFollowedRequest>
                    for CheckUserFollowedSvc<T> {
                        type Response = super::CheckUserFollowedResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckUserFollowedRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::check_user_followed(&inner, request)
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
                        let method = CheckUserFollowedSvc(inner);
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
                "/relation.api.v1.RelationService/PageGetUserFanList" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetUserFanListSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::PageGetUserFanListRequest>
                    for PageGetUserFanListSvc<T> {
                        type Response = super::PageGetUserFanListResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetUserFanListRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::page_get_user_fan_list(
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
                        let method = PageGetUserFanListSvc(inner);
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
                "/relation.api.v1.RelationService/PageGetUserFollowingList" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetUserFollowingListSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::PageGetUserFollowingListRequest>
                    for PageGetUserFollowingListSvc<T> {
                        type Response = super::PageGetUserFollowingListResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::PageGetUserFollowingListRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::page_get_user_following_list(
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
                        let method = PageGetUserFollowingListSvc(inner);
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
                "/relation.api.v1.RelationService/UpdateUserSettings" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateUserSettingsSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::UpdateUserSettingsRequest>
                    for UpdateUserSettingsSvc<T> {
                        type Response = super::UpdateUserSettingsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateUserSettingsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::update_user_settings(
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
                        let method = UpdateUserSettingsSvc(inner);
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
                "/relation.api.v1.RelationService/GetUserSettings" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserSettingsSvc<T: RelationService>(pub Arc<T>);
                    impl<
                        T: RelationService,
                    > tonic::server::UnaryService<super::GetUserSettingsRequest>
                    for GetUserSettingsSvc<T> {
                        type Response = super::GetUserSettingsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserSettingsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as RelationService>::get_user_settings(&inner, request)
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
                        let method = GetUserSettingsSvc(inner);
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
    impl<T> Clone for RelationServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "relation.api.v1.RelationService";
    impl<T> tonic::server::NamedService for RelationServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
