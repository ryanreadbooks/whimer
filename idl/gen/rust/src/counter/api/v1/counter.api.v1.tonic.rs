// @generated
/// Generated client implementations.
pub mod counter_service_client {
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
    pub struct CounterServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl CounterServiceClient<tonic::transport::Channel> {
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
    impl<T> CounterServiceClient<T>
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
        ) -> CounterServiceClient<InterceptedService<T, F>>
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
            CounterServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** 添加一条计数记录
*/
        pub async fn add_record(
            &mut self,
            request: impl tonic::IntoRequest<super::AddRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::AddRecordResponse>,
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
                "/counter.api.v1.CounterService/AddRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("counter.api.v1.CounterService", "AddRecord"));
            self.inner.unary(req, path, codec).await
        }
        /** 取消计数记录
*/
        pub async fn cancel_record(
            &mut self,
            request: impl tonic::IntoRequest<super::CancelRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CancelRecordResponse>,
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
                "/counter.api.v1.CounterService/CancelRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("counter.api.v1.CounterService", "CancelRecord"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取一条计数记录
*/
        pub async fn get_record(
            &mut self,
            request: impl tonic::IntoRequest<super::GetRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetRecordResponse>,
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
                "/counter.api.v1.CounterService/GetRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("counter.api.v1.CounterService", "GetRecord"));
            self.inner.unary(req, path, codec).await
        }
        /** 批量获取计数记录
*/
        pub async fn batch_get_record(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchGetRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetRecordResponse>,
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
                "/counter.api.v1.CounterService/BatchGetRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("counter.api.v1.CounterService", "BatchGetRecord"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取oid计数总数
*/
        pub async fn get_summary(
            &mut self,
            request: impl tonic::IntoRequest<super::GetSummaryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetSummaryResponse>,
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
                "/counter.api.v1.CounterService/GetSummary",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("counter.api.v1.CounterService", "GetSummary"));
            self.inner.unary(req, path, codec).await
        }
        /** 批量获取oid计数总数
*/
        pub async fn batch_get_summary(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchGetSummaryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetSummaryResponse>,
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
                "/counter.api.v1.CounterService/BatchGetSummary",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("counter.api.v1.CounterService", "BatchGetSummary"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 分页获取用户的计数(ActDo)记录
*/
        pub async fn page_get_user_record(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetUserRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserRecordResponse>,
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
                "/counter.api.v1.CounterService/PageGetUserRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("counter.api.v1.CounterService", "PageGetUserRecord"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取一条(ActDo)计数记录
*/
        pub async fn check_has_act_do(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckHasActDoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckHasActDoResponse>,
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
                "/counter.api.v1.CounterService/CheckHasActDo",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("counter.api.v1.CounterService", "CheckHasActDo"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 批量获取(ActDo)计数记录
*/
        pub async fn batch_check_has_act_do(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckHasActDoDoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckHasActDoResponse>,
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
                "/counter.api.v1.CounterService/BatchCheckHasActDo",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "counter.api.v1.CounterService",
                        "BatchCheckHasActDo",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod counter_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with CounterServiceServer.
    #[async_trait]
    pub trait CounterService: std::marker::Send + std::marker::Sync + 'static {
        /** 添加一条计数记录
*/
        async fn add_record(
            &self,
            request: tonic::Request<super::AddRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::AddRecordResponse>,
            tonic::Status,
        >;
        /** 取消计数记录
*/
        async fn cancel_record(
            &self,
            request: tonic::Request<super::CancelRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CancelRecordResponse>,
            tonic::Status,
        >;
        /** 获取一条计数记录
*/
        async fn get_record(
            &self,
            request: tonic::Request<super::GetRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetRecordResponse>,
            tonic::Status,
        >;
        /** 批量获取计数记录
*/
        async fn batch_get_record(
            &self,
            request: tonic::Request<super::BatchGetRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetRecordResponse>,
            tonic::Status,
        >;
        /** 获取oid计数总数
*/
        async fn get_summary(
            &self,
            request: tonic::Request<super::GetSummaryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetSummaryResponse>,
            tonic::Status,
        >;
        /** 批量获取oid计数总数
*/
        async fn batch_get_summary(
            &self,
            request: tonic::Request<super::BatchGetSummaryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetSummaryResponse>,
            tonic::Status,
        >;
        /** 分页获取用户的计数(ActDo)记录
*/
        async fn page_get_user_record(
            &self,
            request: tonic::Request<super::PageGetUserRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetUserRecordResponse>,
            tonic::Status,
        >;
        /** 获取一条(ActDo)计数记录
*/
        async fn check_has_act_do(
            &self,
            request: tonic::Request<super::CheckHasActDoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckHasActDoResponse>,
            tonic::Status,
        >;
        /** 批量获取(ActDo)计数记录
*/
        async fn batch_check_has_act_do(
            &self,
            request: tonic::Request<super::BatchCheckHasActDoDoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckHasActDoResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct CounterServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> CounterServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for CounterServiceServer<T>
    where
        T: CounterService,
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
                "/counter.api.v1.CounterService/AddRecord" => {
                    #[allow(non_camel_case_types)]
                    struct AddRecordSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::AddRecordRequest>
                    for AddRecordSvc<T> {
                        type Response = super::AddRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::AddRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::add_record(&inner, request).await
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
                        let method = AddRecordSvc(inner);
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
                "/counter.api.v1.CounterService/CancelRecord" => {
                    #[allow(non_camel_case_types)]
                    struct CancelRecordSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::CancelRecordRequest>
                    for CancelRecordSvc<T> {
                        type Response = super::CancelRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CancelRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::cancel_record(&inner, request).await
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
                        let method = CancelRecordSvc(inner);
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
                "/counter.api.v1.CounterService/GetRecord" => {
                    #[allow(non_camel_case_types)]
                    struct GetRecordSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::GetRecordRequest>
                    for GetRecordSvc<T> {
                        type Response = super::GetRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::get_record(&inner, request).await
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
                        let method = GetRecordSvc(inner);
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
                "/counter.api.v1.CounterService/BatchGetRecord" => {
                    #[allow(non_camel_case_types)]
                    struct BatchGetRecordSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::BatchGetRecordRequest>
                    for BatchGetRecordSvc<T> {
                        type Response = super::BatchGetRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchGetRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::batch_get_record(&inner, request)
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
                        let method = BatchGetRecordSvc(inner);
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
                "/counter.api.v1.CounterService/GetSummary" => {
                    #[allow(non_camel_case_types)]
                    struct GetSummarySvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::GetSummaryRequest>
                    for GetSummarySvc<T> {
                        type Response = super::GetSummaryResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetSummaryRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::get_summary(&inner, request).await
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
                        let method = GetSummarySvc(inner);
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
                "/counter.api.v1.CounterService/BatchGetSummary" => {
                    #[allow(non_camel_case_types)]
                    struct BatchGetSummarySvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::BatchGetSummaryRequest>
                    for BatchGetSummarySvc<T> {
                        type Response = super::BatchGetSummaryResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchGetSummaryRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::batch_get_summary(&inner, request)
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
                        let method = BatchGetSummarySvc(inner);
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
                "/counter.api.v1.CounterService/PageGetUserRecord" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetUserRecordSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::PageGetUserRecordRequest>
                    for PageGetUserRecordSvc<T> {
                        type Response = super::PageGetUserRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetUserRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::page_get_user_record(&inner, request)
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
                        let method = PageGetUserRecordSvc(inner);
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
                "/counter.api.v1.CounterService/CheckHasActDo" => {
                    #[allow(non_camel_case_types)]
                    struct CheckHasActDoSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::CheckHasActDoRequest>
                    for CheckHasActDoSvc<T> {
                        type Response = super::CheckHasActDoResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckHasActDoRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::check_has_act_do(&inner, request)
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
                        let method = CheckHasActDoSvc(inner);
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
                "/counter.api.v1.CounterService/BatchCheckHasActDo" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckHasActDoSvc<T: CounterService>(pub Arc<T>);
                    impl<
                        T: CounterService,
                    > tonic::server::UnaryService<super::BatchCheckHasActDoDoRequest>
                    for BatchCheckHasActDoSvc<T> {
                        type Response = super::BatchCheckHasActDoResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchCheckHasActDoDoRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CounterService>::batch_check_has_act_do(
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
                        let method = BatchCheckHasActDoSvc(inner);
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
    impl<T> Clone for CounterServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "counter.api.v1.CounterService";
    impl<T> tonic::server::NamedService for CounterServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
