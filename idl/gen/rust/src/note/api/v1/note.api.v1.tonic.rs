// @generated
/// Generated client implementations.
pub mod note_creator_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** 和笔记管理相关的服务
 比如发布笔记，修改笔记，删除笔记等管理笔记的功能
*/
    #[derive(Debug, Clone)]
    pub struct NoteCreatorServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl NoteCreatorServiceClient<tonic::transport::Channel> {
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
    impl<T> NoteCreatorServiceClient<T>
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
        ) -> NoteCreatorServiceClient<InterceptedService<T, F>>
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
            NoteCreatorServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** 检查用户是否拥有指定的笔记
*/
        pub async fn is_user_own_note(
            &mut self,
            request: impl tonic::IntoRequest<super::IsUserOwnNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::IsUserOwnNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/IsUserOwnNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteCreatorService", "IsUserOwnNote"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 判断笔记是否存在
*/
        pub async fn is_note_exist(
            &mut self,
            request: impl tonic::IntoRequest<super::IsNoteExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::IsNoteExistResponse>,
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
                "/note.api.v1.NoteCreatorService/IsNoteExist",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteCreatorService", "IsNoteExist"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 创建笔记
*/
        pub async fn create_note(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/CreateNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "CreateNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 更新笔记
*/
        pub async fn update_note(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/UpdateNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "UpdateNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 删除笔记
*/
        pub async fn delete_note(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/DeleteNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "DeleteNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取笔记的信息
*/
        pub async fn get_note(
            &mut self,
            request: impl tonic::IntoRequest<super::GetNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/GetNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "GetNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 列出笔记
*/
        pub async fn list_note(
            &mut self,
            request: impl tonic::IntoRequest<super::ListNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/ListNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "ListNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 分页列出笔记
*/
        pub async fn page_list_note(
            &mut self,
            request: impl tonic::IntoRequest<super::PageListNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageListNoteResponse>,
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
                "/note.api.v1.NoteCreatorService/PageListNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteCreatorService", "PageListNote"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取用户投稿数量
*/
        pub async fn get_posted_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetPostedCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPostedCountResponse>,
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
                "/note.api.v1.NoteCreatorService/GetPostedCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteCreatorService", "GetPostedCount"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 新增标签
*/
        pub async fn add_tag(
            &mut self,
            request: impl tonic::IntoRequest<super::AddTagRequest>,
        ) -> std::result::Result<tonic::Response<super::AddTagResponse>, tonic::Status> {
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
                "/note.api.v1.NoteCreatorService/AddTag",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteCreatorService", "AddTag"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod note_creator_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with NoteCreatorServiceServer.
    #[async_trait]
    pub trait NoteCreatorService: std::marker::Send + std::marker::Sync + 'static {
        /** 检查用户是否拥有指定的笔记
*/
        async fn is_user_own_note(
            &self,
            request: tonic::Request<super::IsUserOwnNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::IsUserOwnNoteResponse>,
            tonic::Status,
        >;
        /** 判断笔记是否存在
*/
        async fn is_note_exist(
            &self,
            request: tonic::Request<super::IsNoteExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::IsNoteExistResponse>,
            tonic::Status,
        >;
        /** 创建笔记
*/
        async fn create_note(
            &self,
            request: tonic::Request<super::CreateNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateNoteResponse>,
            tonic::Status,
        >;
        /** 更新笔记
*/
        async fn update_note(
            &self,
            request: tonic::Request<super::UpdateNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateNoteResponse>,
            tonic::Status,
        >;
        /** 删除笔记
*/
        async fn delete_note(
            &self,
            request: tonic::Request<super::DeleteNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteNoteResponse>,
            tonic::Status,
        >;
        /** 获取笔记的信息
*/
        async fn get_note(
            &self,
            request: tonic::Request<super::GetNoteRequest>,
        ) -> std::result::Result<tonic::Response<super::GetNoteResponse>, tonic::Status>;
        /** 列出笔记
*/
        async fn list_note(
            &self,
            request: tonic::Request<super::ListNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListNoteResponse>,
            tonic::Status,
        >;
        /** 分页列出笔记
*/
        async fn page_list_note(
            &self,
            request: tonic::Request<super::PageListNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageListNoteResponse>,
            tonic::Status,
        >;
        /** 获取用户投稿数量
*/
        async fn get_posted_count(
            &self,
            request: tonic::Request<super::GetPostedCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPostedCountResponse>,
            tonic::Status,
        >;
        /** 新增标签
*/
        async fn add_tag(
            &self,
            request: tonic::Request<super::AddTagRequest>,
        ) -> std::result::Result<tonic::Response<super::AddTagResponse>, tonic::Status>;
    }
    /** 和笔记管理相关的服务
 比如发布笔记，修改笔记，删除笔记等管理笔记的功能
*/
    #[derive(Debug)]
    pub struct NoteCreatorServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> NoteCreatorServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for NoteCreatorServiceServer<T>
    where
        T: NoteCreatorService,
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
                "/note.api.v1.NoteCreatorService/IsUserOwnNote" => {
                    #[allow(non_camel_case_types)]
                    struct IsUserOwnNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::IsUserOwnNoteRequest>
                    for IsUserOwnNoteSvc<T> {
                        type Response = super::IsUserOwnNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::IsUserOwnNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::is_user_own_note(&inner, request)
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
                        let method = IsUserOwnNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/IsNoteExist" => {
                    #[allow(non_camel_case_types)]
                    struct IsNoteExistSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::IsNoteExistRequest>
                    for IsNoteExistSvc<T> {
                        type Response = super::IsNoteExistResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::IsNoteExistRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::is_note_exist(&inner, request)
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
                        let method = IsNoteExistSvc(inner);
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
                "/note.api.v1.NoteCreatorService/CreateNote" => {
                    #[allow(non_camel_case_types)]
                    struct CreateNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::CreateNoteRequest>
                    for CreateNoteSvc<T> {
                        type Response = super::CreateNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::create_note(&inner, request)
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
                        let method = CreateNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/UpdateNote" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::UpdateNoteRequest>
                    for UpdateNoteSvc<T> {
                        type Response = super::UpdateNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::update_note(&inner, request)
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
                        let method = UpdateNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/DeleteNote" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::DeleteNoteRequest>
                    for DeleteNoteSvc<T> {
                        type Response = super::DeleteNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::delete_note(&inner, request)
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
                        let method = DeleteNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/GetNote" => {
                    #[allow(non_camel_case_types)]
                    struct GetNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::GetNoteRequest>
                    for GetNoteSvc<T> {
                        type Response = super::GetNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::get_note(&inner, request).await
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
                        let method = GetNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/ListNote" => {
                    #[allow(non_camel_case_types)]
                    struct ListNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::ListNoteRequest>
                    for ListNoteSvc<T> {
                        type Response = super::ListNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::list_note(&inner, request).await
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
                        let method = ListNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/PageListNote" => {
                    #[allow(non_camel_case_types)]
                    struct PageListNoteSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::PageListNoteRequest>
                    for PageListNoteSvc<T> {
                        type Response = super::PageListNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageListNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::page_list_note(&inner, request)
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
                        let method = PageListNoteSvc(inner);
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
                "/note.api.v1.NoteCreatorService/GetPostedCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetPostedCountSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::GetPostedCountRequest>
                    for GetPostedCountSvc<T> {
                        type Response = super::GetPostedCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetPostedCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::get_posted_count(&inner, request)
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
                        let method = GetPostedCountSvc(inner);
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
                "/note.api.v1.NoteCreatorService/AddTag" => {
                    #[allow(non_camel_case_types)]
                    struct AddTagSvc<T: NoteCreatorService>(pub Arc<T>);
                    impl<
                        T: NoteCreatorService,
                    > tonic::server::UnaryService<super::AddTagRequest>
                    for AddTagSvc<T> {
                        type Response = super::AddTagResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::AddTagRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteCreatorService>::add_tag(&inner, request).await
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
                        let method = AddTagSvc(inner);
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
    impl<T> Clone for NoteCreatorServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "note.api.v1.NoteCreatorService";
    impl<T> tonic::server::NamedService for NoteCreatorServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
/// Generated client implementations.
pub mod note_interact_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** 与笔记交互逻辑相关服务，比如点赞、收藏等
*/
    #[derive(Debug, Clone)]
    pub struct NoteInteractServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl NoteInteractServiceClient<tonic::transport::Channel> {
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
    impl<T> NoteInteractServiceClient<T>
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
        ) -> NoteInteractServiceClient<InterceptedService<T, F>>
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
            NoteInteractServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** 点赞笔记/取消点赞
*/
        pub async fn like_note(
            &mut self,
            request: impl tonic::IntoRequest<super::LikeNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LikeNoteResponse>,
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
                "/note.api.v1.NoteInteractService/LikeNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteInteractService", "LikeNote"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取笔记点赞数量
*/
        pub async fn get_note_likes(
            &mut self,
            request: impl tonic::IntoRequest<super::GetNoteLikesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteLikesResponse>,
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
                "/note.api.v1.NoteInteractService/GetNoteLikes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteInteractService", "GetNoteLikes"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 检查某个用户是否点赞过某篇笔记
*/
        pub async fn check_user_like_status(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckUserLikeStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserLikeStatusResponse>,
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
                "/note.api.v1.NoteInteractService/CheckUserLikeStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteInteractService",
                        "CheckUserLikeStatus",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 批量检查用户是否点赞过多篇笔记
*/
        pub async fn batch_check_user_like_status(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckUserLikeStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserLikeStatusResponse>,
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
                "/note.api.v1.NoteInteractService/BatchCheckUserLikeStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteInteractService",
                        "BatchCheckUserLikeStatus",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取笔记的交互信息
*/
        pub async fn get_note_interaction(
            &mut self,
            request: impl tonic::IntoRequest<super::GetNoteInteractionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteInteractionResponse>,
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
                "/note.api.v1.NoteInteractService/GetNoteInteraction",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteInteractService",
                        "GetNoteInteraction",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取用户点赞过的笔记
*/
        pub async fn page_list_user_liked_note(
            &mut self,
            request: impl tonic::IntoRequest<super::PageListUserLikedNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageListUserLikedNoteResponse>,
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
                "/note.api.v1.NoteInteractService/PageListUserLikedNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteInteractService",
                        "PageListUserLikedNote",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod note_interact_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with NoteInteractServiceServer.
    #[async_trait]
    pub trait NoteInteractService: std::marker::Send + std::marker::Sync + 'static {
        /** 点赞笔记/取消点赞
*/
        async fn like_note(
            &self,
            request: tonic::Request<super::LikeNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LikeNoteResponse>,
            tonic::Status,
        >;
        /** 获取笔记点赞数量
*/
        async fn get_note_likes(
            &self,
            request: tonic::Request<super::GetNoteLikesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteLikesResponse>,
            tonic::Status,
        >;
        /** 检查某个用户是否点赞过某篇笔记
*/
        async fn check_user_like_status(
            &self,
            request: tonic::Request<super::CheckUserLikeStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserLikeStatusResponse>,
            tonic::Status,
        >;
        /** 批量检查用户是否点赞过多篇笔记
*/
        async fn batch_check_user_like_status(
            &self,
            request: tonic::Request<super::BatchCheckUserLikeStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserLikeStatusResponse>,
            tonic::Status,
        >;
        /** 获取笔记的交互信息
*/
        async fn get_note_interaction(
            &self,
            request: tonic::Request<super::GetNoteInteractionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteInteractionResponse>,
            tonic::Status,
        >;
        /** 获取用户点赞过的笔记
*/
        async fn page_list_user_liked_note(
            &self,
            request: tonic::Request<super::PageListUserLikedNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageListUserLikedNoteResponse>,
            tonic::Status,
        >;
    }
    /** 与笔记交互逻辑相关服务，比如点赞、收藏等
*/
    #[derive(Debug)]
    pub struct NoteInteractServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> NoteInteractServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for NoteInteractServiceServer<T>
    where
        T: NoteInteractService,
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
                "/note.api.v1.NoteInteractService/LikeNote" => {
                    #[allow(non_camel_case_types)]
                    struct LikeNoteSvc<T: NoteInteractService>(pub Arc<T>);
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::LikeNoteRequest>
                    for LikeNoteSvc<T> {
                        type Response = super::LikeNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::LikeNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::like_note(&inner, request).await
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
                        let method = LikeNoteSvc(inner);
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
                "/note.api.v1.NoteInteractService/GetNoteLikes" => {
                    #[allow(non_camel_case_types)]
                    struct GetNoteLikesSvc<T: NoteInteractService>(pub Arc<T>);
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::GetNoteLikesRequest>
                    for GetNoteLikesSvc<T> {
                        type Response = super::GetNoteLikesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetNoteLikesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::get_note_likes(&inner, request)
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
                        let method = GetNoteLikesSvc(inner);
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
                "/note.api.v1.NoteInteractService/CheckUserLikeStatus" => {
                    #[allow(non_camel_case_types)]
                    struct CheckUserLikeStatusSvc<T: NoteInteractService>(pub Arc<T>);
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::CheckUserLikeStatusRequest>
                    for CheckUserLikeStatusSvc<T> {
                        type Response = super::CheckUserLikeStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckUserLikeStatusRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::check_user_like_status(
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
                        let method = CheckUserLikeStatusSvc(inner);
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
                "/note.api.v1.NoteInteractService/BatchCheckUserLikeStatus" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckUserLikeStatusSvc<T: NoteInteractService>(
                        pub Arc<T>,
                    );
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::BatchCheckUserLikeStatusRequest>
                    for BatchCheckUserLikeStatusSvc<T> {
                        type Response = super::BatchCheckUserLikeStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::BatchCheckUserLikeStatusRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::batch_check_user_like_status(
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
                        let method = BatchCheckUserLikeStatusSvc(inner);
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
                "/note.api.v1.NoteInteractService/GetNoteInteraction" => {
                    #[allow(non_camel_case_types)]
                    struct GetNoteInteractionSvc<T: NoteInteractService>(pub Arc<T>);
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::GetNoteInteractionRequest>
                    for GetNoteInteractionSvc<T> {
                        type Response = super::GetNoteInteractionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetNoteInteractionRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::get_note_interaction(
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
                        let method = GetNoteInteractionSvc(inner);
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
                "/note.api.v1.NoteInteractService/PageListUserLikedNote" => {
                    #[allow(non_camel_case_types)]
                    struct PageListUserLikedNoteSvc<T: NoteInteractService>(pub Arc<T>);
                    impl<
                        T: NoteInteractService,
                    > tonic::server::UnaryService<super::PageListUserLikedNoteRequest>
                    for PageListUserLikedNoteSvc<T> {
                        type Response = super::PageListUserLikedNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageListUserLikedNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteInteractService>::page_list_user_liked_note(
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
                        let method = PageListUserLikedNoteSvc(inner);
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
    impl<T> Clone for NoteInteractServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "note.api.v1.NoteInteractService";
    impl<T> tonic::server::NamedService for NoteInteractServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
/// Generated client implementations.
pub mod note_feed_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** note相关非管理功能服务
*/
    #[derive(Debug, Clone)]
    pub struct NoteFeedServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl NoteFeedServiceClient<tonic::transport::Channel> {
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
    impl<T> NoteFeedServiceClient<T>
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
        ) -> NoteFeedServiceClient<InterceptedService<T, F>>
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
            NoteFeedServiceClient::new(InterceptedService::new(inner, interceptor))
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
        pub async fn random_get(
            &mut self,
            request: impl tonic::IntoRequest<super::RandomGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RandomGetResponse>,
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
                "/note.api.v1.NoteFeedService/RandomGet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "RandomGet"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_feed_note(
            &mut self,
            request: impl tonic::IntoRequest<super::GetFeedNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetFeedNoteResponse>,
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
                "/note.api.v1.NoteFeedService/GetFeedNote",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "GetFeedNote"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn get_note_author(
            &mut self,
            request: impl tonic::IntoRequest<super::GetNoteAuthorRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteAuthorResponse>,
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
                "/note.api.v1.NoteFeedService/GetNoteAuthor",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "GetNoteAuthor"));
            self.inner.unary(req, path, codec).await
        }
        /** 批量获取笔记
*/
        pub async fn batch_get_feed_notes(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchGetFeedNotesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetFeedNotesResponse>,
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
                "/note.api.v1.NoteFeedService/BatchGetFeedNotes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteFeedService", "BatchGetFeedNotes"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 按照推荐获取
*/
        pub async fn recommend_get(
            &mut self,
            request: impl tonic::IntoRequest<super::RecommendGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RecommendGetResponse>,
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
                "/note.api.v1.NoteFeedService/RecommendGet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "RecommendGet"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取指定用户的最近的笔记内容
*/
        pub async fn get_user_recent_post(
            &mut self,
            request: impl tonic::IntoRequest<super::GetUserRecentPostRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserRecentPostResponse>,
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
                "/note.api.v1.NoteFeedService/GetUserRecentPost",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("note.api.v1.NoteFeedService", "GetUserRecentPost"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 列出指定用户公开的笔记内容
*/
        pub async fn list_feed_by_uid(
            &mut self,
            request: impl tonic::IntoRequest<super::ListFeedByUidRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListFeedByUidResponse>,
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
                "/note.api.v1.NoteFeedService/ListFeedByUid",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "ListFeedByUid"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取笔记标签
*/
        pub async fn get_tag_info(
            &mut self,
            request: impl tonic::IntoRequest<super::GetTagInfoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetTagInfoResponse>,
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
                "/note.api.v1.NoteFeedService/GetTagInfo",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("note.api.v1.NoteFeedService", "GetTagInfo"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取用户投稿数量
*/
        pub async fn get_public_posted_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetPublicPostedCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPublicPostedCountResponse>,
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
                "/note.api.v1.NoteFeedService/GetPublicPostedCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteFeedService",
                        "GetPublicPostedCount",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 检查笔记是否存在（存在且公开）
*/
        pub async fn batch_check_feed_note_exist(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckFeedNoteExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckFeedNoteExistResponse>,
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
                "/note.api.v1.NoteFeedService/BatchCheckFeedNoteExist",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "note.api.v1.NoteFeedService",
                        "BatchCheckFeedNoteExist",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod note_feed_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with NoteFeedServiceServer.
    #[async_trait]
    pub trait NoteFeedService: std::marker::Send + std::marker::Sync + 'static {
        async fn random_get(
            &self,
            request: tonic::Request<super::RandomGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RandomGetResponse>,
            tonic::Status,
        >;
        async fn get_feed_note(
            &self,
            request: tonic::Request<super::GetFeedNoteRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetFeedNoteResponse>,
            tonic::Status,
        >;
        async fn get_note_author(
            &self,
            request: tonic::Request<super::GetNoteAuthorRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetNoteAuthorResponse>,
            tonic::Status,
        >;
        /** 批量获取笔记
*/
        async fn batch_get_feed_notes(
            &self,
            request: tonic::Request<super::BatchGetFeedNotesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchGetFeedNotesResponse>,
            tonic::Status,
        >;
        /** 按照推荐获取
*/
        async fn recommend_get(
            &self,
            request: tonic::Request<super::RecommendGetRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RecommendGetResponse>,
            tonic::Status,
        >;
        /** 获取指定用户的最近的笔记内容
*/
        async fn get_user_recent_post(
            &self,
            request: tonic::Request<super::GetUserRecentPostRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetUserRecentPostResponse>,
            tonic::Status,
        >;
        /** 列出指定用户公开的笔记内容
*/
        async fn list_feed_by_uid(
            &self,
            request: tonic::Request<super::ListFeedByUidRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListFeedByUidResponse>,
            tonic::Status,
        >;
        /** 获取笔记标签
*/
        async fn get_tag_info(
            &self,
            request: tonic::Request<super::GetTagInfoRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetTagInfoResponse>,
            tonic::Status,
        >;
        /** 获取用户投稿数量
*/
        async fn get_public_posted_count(
            &self,
            request: tonic::Request<super::GetPublicPostedCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPublicPostedCountResponse>,
            tonic::Status,
        >;
        /** 检查笔记是否存在（存在且公开）
*/
        async fn batch_check_feed_note_exist(
            &self,
            request: tonic::Request<super::BatchCheckFeedNoteExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckFeedNoteExistResponse>,
            tonic::Status,
        >;
    }
    /** note相关非管理功能服务
*/
    #[derive(Debug)]
    pub struct NoteFeedServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> NoteFeedServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for NoteFeedServiceServer<T>
    where
        T: NoteFeedService,
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
                "/note.api.v1.NoteFeedService/RandomGet" => {
                    #[allow(non_camel_case_types)]
                    struct RandomGetSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::RandomGetRequest>
                    for RandomGetSvc<T> {
                        type Response = super::RandomGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RandomGetRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::random_get(&inner, request).await
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
                        let method = RandomGetSvc(inner);
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
                "/note.api.v1.NoteFeedService/GetFeedNote" => {
                    #[allow(non_camel_case_types)]
                    struct GetFeedNoteSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::GetFeedNoteRequest>
                    for GetFeedNoteSvc<T> {
                        type Response = super::GetFeedNoteResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetFeedNoteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::get_feed_note(&inner, request).await
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
                        let method = GetFeedNoteSvc(inner);
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
                "/note.api.v1.NoteFeedService/GetNoteAuthor" => {
                    #[allow(non_camel_case_types)]
                    struct GetNoteAuthorSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::GetNoteAuthorRequest>
                    for GetNoteAuthorSvc<T> {
                        type Response = super::GetNoteAuthorResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetNoteAuthorRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::get_note_author(&inner, request)
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
                        let method = GetNoteAuthorSvc(inner);
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
                "/note.api.v1.NoteFeedService/BatchGetFeedNotes" => {
                    #[allow(non_camel_case_types)]
                    struct BatchGetFeedNotesSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::BatchGetFeedNotesRequest>
                    for BatchGetFeedNotesSvc<T> {
                        type Response = super::BatchGetFeedNotesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchGetFeedNotesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::batch_get_feed_notes(
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
                        let method = BatchGetFeedNotesSvc(inner);
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
                "/note.api.v1.NoteFeedService/RecommendGet" => {
                    #[allow(non_camel_case_types)]
                    struct RecommendGetSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::RecommendGetRequest>
                    for RecommendGetSvc<T> {
                        type Response = super::RecommendGetResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RecommendGetRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::recommend_get(&inner, request).await
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
                        let method = RecommendGetSvc(inner);
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
                "/note.api.v1.NoteFeedService/GetUserRecentPost" => {
                    #[allow(non_camel_case_types)]
                    struct GetUserRecentPostSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::GetUserRecentPostRequest>
                    for GetUserRecentPostSvc<T> {
                        type Response = super::GetUserRecentPostResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetUserRecentPostRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::get_user_recent_post(
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
                        let method = GetUserRecentPostSvc(inner);
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
                "/note.api.v1.NoteFeedService/ListFeedByUid" => {
                    #[allow(non_camel_case_types)]
                    struct ListFeedByUidSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::ListFeedByUidRequest>
                    for ListFeedByUidSvc<T> {
                        type Response = super::ListFeedByUidResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListFeedByUidRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::list_feed_by_uid(&inner, request)
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
                        let method = ListFeedByUidSvc(inner);
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
                "/note.api.v1.NoteFeedService/GetTagInfo" => {
                    #[allow(non_camel_case_types)]
                    struct GetTagInfoSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::GetTagInfoRequest>
                    for GetTagInfoSvc<T> {
                        type Response = super::GetTagInfoResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetTagInfoRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::get_tag_info(&inner, request).await
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
                        let method = GetTagInfoSvc(inner);
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
                "/note.api.v1.NoteFeedService/GetPublicPostedCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetPublicPostedCountSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::GetPublicPostedCountRequest>
                    for GetPublicPostedCountSvc<T> {
                        type Response = super::GetPublicPostedCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetPublicPostedCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::get_public_posted_count(
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
                        let method = GetPublicPostedCountSvc(inner);
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
                "/note.api.v1.NoteFeedService/BatchCheckFeedNoteExist" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckFeedNoteExistSvc<T: NoteFeedService>(pub Arc<T>);
                    impl<
                        T: NoteFeedService,
                    > tonic::server::UnaryService<super::BatchCheckFeedNoteExistRequest>
                    for BatchCheckFeedNoteExistSvc<T> {
                        type Response = super::BatchCheckFeedNoteExistResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::BatchCheckFeedNoteExistRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as NoteFeedService>::batch_check_feed_note_exist(
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
                        let method = BatchCheckFeedNoteExistSvc(inner);
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
    impl<T> Clone for NoteFeedServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "note.api.v1.NoteFeedService";
    impl<T> tonic::server::NamedService for NoteFeedServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
