// @generated
/// Generated client implementations.
pub mod comment_service_client {
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
    pub struct CommentServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl CommentServiceClient<tonic::transport::Channel> {
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
    impl<T> CommentServiceClient<T>
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
        ) -> CommentServiceClient<InterceptedService<T, F>>
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
            CommentServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** 发表评论
*/
        pub async fn add_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::AddCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::AddCommentResponse>,
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
                "/comment.api.v1.CommentService/AddComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("comment.api.v1.CommentService", "AddComment"));
            self.inner.unary(req, path, codec).await
        }
        /** 删除评论
*/
        pub async fn del_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::DelCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DelCommentResponse>,
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
                "/comment.api.v1.CommentService/DelComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("comment.api.v1.CommentService", "DelComment"));
            self.inner.unary(req, path, codec).await
        }
        /** 赞
*/
        pub async fn like_action(
            &mut self,
            request: impl tonic::IntoRequest<super::LikeActionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LikeActionResponse>,
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
                "/comment.api.v1.CommentService/LikeAction",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("comment.api.v1.CommentService", "LikeAction"));
            self.inner.unary(req, path, codec).await
        }
        /** 踩
*/
        pub async fn dislike_action(
            &mut self,
            request: impl tonic::IntoRequest<super::DislikeActionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DislikeActionResponse>,
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
                "/comment.api.v1.CommentService/DislikeAction",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "DislikeAction"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 举报
*/
        pub async fn report_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::ReportCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ReportCommentResponse>,
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
                "/comment.api.v1.CommentService/ReportComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "ReportComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 置顶评论
*/
        pub async fn pin_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::PinCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PinCommentResponse>,
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
                "/comment.api.v1.CommentService/PinComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("comment.api.v1.CommentService", "PinComment"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取主评论信息
*/
        pub async fn page_get_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetCommentResponse>,
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
                "/comment.api.v1.CommentService/PageGetComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "PageGetComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取子评论信息
*/
        pub async fn page_get_sub_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetSubCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetSubCommentResponse>,
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
                "/comment.api.v1.CommentService/PageGetSubComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "PageGetSubComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 分页获取子评论信息
*/
        pub async fn page_get_sub_comment_v2(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetSubCommentV2Request>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetSubCommentV2Response>,
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
                "/comment.api.v1.CommentService/PageGetSubCommentV2",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "PageGetSubCommentV2",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取主评论详细信息
*/
        pub async fn page_get_detailed_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetDetailedCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetDetailedCommentResponse>,
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
                "/comment.api.v1.CommentService/PageGetDetailedComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "PageGetDetailedComment",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn page_get_detailed_comment_v2(
            &mut self,
            request: impl tonic::IntoRequest<super::PageGetDetailedCommentV2Request>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetDetailedCommentV2Response>,
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
                "/comment.api.v1.CommentService/PageGetDetailedCommentV2",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "PageGetDetailedCommentV2",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取置顶评论
*/
        pub async fn get_pinned_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::GetPinnedCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPinnedCommentResponse>,
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
                "/comment.api.v1.CommentService/GetPinnedComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "GetPinnedComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某个被评对象的评论数
*/
        pub async fn count_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::CountCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CountCommentResponse>,
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
                "/comment.api.v1.CommentService/CountComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "CountComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取多个被评论对象的评论数
*/
        pub async fn batch_count_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCountCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCountCommentResponse>,
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
                "/comment.api.v1.CommentService/BatchCountComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "BatchCountComment"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某条评论的点赞数
*/
        pub async fn get_comment_like_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommentLikeCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentLikeCountResponse>,
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
                "/comment.api.v1.CommentService/GetCommentLikeCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "GetCommentLikeCount",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某条评论的点踩数
*/
        pub async fn get_comment_dislike_count(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommentDislikeCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentDislikeCountResponse>,
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
                "/comment.api.v1.CommentService/GetCommentDislikeCount",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "GetCommentDislikeCount",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 获取某个用户是否评论了某个对象
*/
        pub async fn check_user_on_object(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckUserOnObjectRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserOnObjectResponse>,
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
                "/comment.api.v1.CommentService/CheckUserOnObject",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "CheckUserOnObject"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 批量检查某个用户是否评论了某个对象
*/
        pub async fn batch_check_user_on_object(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckUserOnObjectRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserOnObjectResponse>,
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
                "/comment.api.v1.CommentService/BatchCheckUserOnObject",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "BatchCheckUserOnObject",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 批量检查某个用户是否点赞了某些评论
*/
        pub async fn batch_check_user_like_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckUserLikeCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserLikeCommentResponse>,
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
                "/comment.api.v1.CommentService/BatchCheckUserLikeComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "BatchCheckUserLikeComment",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 批量检查评论是否存在
*/
        pub async fn batch_check_comment_exist(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchCheckCommentExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckCommentExistResponse>,
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
                "/comment.api.v1.CommentService/BatchCheckCommentExist",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "comment.api.v1.CommentService",
                        "BatchCheckCommentExist",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** 按照id获取评论
*/
        pub async fn get_comment(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentResponse>,
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
                "/comment.api.v1.CommentService/GetComment",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("comment.api.v1.CommentService", "GetComment"));
            self.inner.unary(req, path, codec).await
        }
        /** 获取评论作者
*/
        pub async fn get_comment_user(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommentUserRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentUserResponse>,
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
                "/comment.api.v1.CommentService/GetCommentUser",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("comment.api.v1.CommentService", "GetCommentUser"),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod comment_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with CommentServiceServer.
    #[async_trait]
    pub trait CommentService: std::marker::Send + std::marker::Sync + 'static {
        /** 发表评论
*/
        async fn add_comment(
            &self,
            request: tonic::Request<super::AddCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::AddCommentResponse>,
            tonic::Status,
        >;
        /** 删除评论
*/
        async fn del_comment(
            &self,
            request: tonic::Request<super::DelCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DelCommentResponse>,
            tonic::Status,
        >;
        /** 赞
*/
        async fn like_action(
            &self,
            request: tonic::Request<super::LikeActionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LikeActionResponse>,
            tonic::Status,
        >;
        /** 踩
*/
        async fn dislike_action(
            &self,
            request: tonic::Request<super::DislikeActionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DislikeActionResponse>,
            tonic::Status,
        >;
        /** 举报
*/
        async fn report_comment(
            &self,
            request: tonic::Request<super::ReportCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ReportCommentResponse>,
            tonic::Status,
        >;
        /** 置顶评论
*/
        async fn pin_comment(
            &self,
            request: tonic::Request<super::PinCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PinCommentResponse>,
            tonic::Status,
        >;
        /** 获取主评论信息
*/
        async fn page_get_comment(
            &self,
            request: tonic::Request<super::PageGetCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetCommentResponse>,
            tonic::Status,
        >;
        /** 获取子评论信息
*/
        async fn page_get_sub_comment(
            &self,
            request: tonic::Request<super::PageGetSubCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetSubCommentResponse>,
            tonic::Status,
        >;
        /** 分页获取子评论信息
*/
        async fn page_get_sub_comment_v2(
            &self,
            request: tonic::Request<super::PageGetSubCommentV2Request>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetSubCommentV2Response>,
            tonic::Status,
        >;
        /** 获取主评论详细信息
*/
        async fn page_get_detailed_comment(
            &self,
            request: tonic::Request<super::PageGetDetailedCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetDetailedCommentResponse>,
            tonic::Status,
        >;
        ///
        async fn page_get_detailed_comment_v2(
            &self,
            request: tonic::Request<super::PageGetDetailedCommentV2Request>,
        ) -> std::result::Result<
            tonic::Response<super::PageGetDetailedCommentV2Response>,
            tonic::Status,
        >;
        /** 获取置顶评论
*/
        async fn get_pinned_comment(
            &self,
            request: tonic::Request<super::GetPinnedCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPinnedCommentResponse>,
            tonic::Status,
        >;
        /** 获取某个被评对象的评论数
*/
        async fn count_comment(
            &self,
            request: tonic::Request<super::CountCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CountCommentResponse>,
            tonic::Status,
        >;
        /** 获取多个被评论对象的评论数
*/
        async fn batch_count_comment(
            &self,
            request: tonic::Request<super::BatchCountCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCountCommentResponse>,
            tonic::Status,
        >;
        /** 获取某条评论的点赞数
*/
        async fn get_comment_like_count(
            &self,
            request: tonic::Request<super::GetCommentLikeCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentLikeCountResponse>,
            tonic::Status,
        >;
        /** 获取某条评论的点踩数
*/
        async fn get_comment_dislike_count(
            &self,
            request: tonic::Request<super::GetCommentDislikeCountRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentDislikeCountResponse>,
            tonic::Status,
        >;
        /** 获取某个用户是否评论了某个对象
*/
        async fn check_user_on_object(
            &self,
            request: tonic::Request<super::CheckUserOnObjectRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckUserOnObjectResponse>,
            tonic::Status,
        >;
        /** 批量检查某个用户是否评论了某个对象
*/
        async fn batch_check_user_on_object(
            &self,
            request: tonic::Request<super::BatchCheckUserOnObjectRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserOnObjectResponse>,
            tonic::Status,
        >;
        /** 批量检查某个用户是否点赞了某些评论
*/
        async fn batch_check_user_like_comment(
            &self,
            request: tonic::Request<super::BatchCheckUserLikeCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckUserLikeCommentResponse>,
            tonic::Status,
        >;
        /** 批量检查评论是否存在
*/
        async fn batch_check_comment_exist(
            &self,
            request: tonic::Request<super::BatchCheckCommentExistRequest>,
        ) -> std::result::Result<
            tonic::Response<super::BatchCheckCommentExistResponse>,
            tonic::Status,
        >;
        /** 按照id获取评论
*/
        async fn get_comment(
            &self,
            request: tonic::Request<super::GetCommentRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentResponse>,
            tonic::Status,
        >;
        /** 获取评论作者
*/
        async fn get_comment_user(
            &self,
            request: tonic::Request<super::GetCommentUserRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommentUserResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct CommentServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> CommentServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for CommentServiceServer<T>
    where
        T: CommentService,
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
                "/comment.api.v1.CommentService/AddComment" => {
                    #[allow(non_camel_case_types)]
                    struct AddCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::AddCommentRequest>
                    for AddCommentSvc<T> {
                        type Response = super::AddCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::AddCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::add_comment(&inner, request).await
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
                        let method = AddCommentSvc(inner);
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
                "/comment.api.v1.CommentService/DelComment" => {
                    #[allow(non_camel_case_types)]
                    struct DelCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::DelCommentRequest>
                    for DelCommentSvc<T> {
                        type Response = super::DelCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DelCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::del_comment(&inner, request).await
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
                        let method = DelCommentSvc(inner);
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
                "/comment.api.v1.CommentService/LikeAction" => {
                    #[allow(non_camel_case_types)]
                    struct LikeActionSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::LikeActionRequest>
                    for LikeActionSvc<T> {
                        type Response = super::LikeActionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::LikeActionRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::like_action(&inner, request).await
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
                        let method = LikeActionSvc(inner);
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
                "/comment.api.v1.CommentService/DislikeAction" => {
                    #[allow(non_camel_case_types)]
                    struct DislikeActionSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::DislikeActionRequest>
                    for DislikeActionSvc<T> {
                        type Response = super::DislikeActionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DislikeActionRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::dislike_action(&inner, request).await
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
                        let method = DislikeActionSvc(inner);
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
                "/comment.api.v1.CommentService/ReportComment" => {
                    #[allow(non_camel_case_types)]
                    struct ReportCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::ReportCommentRequest>
                    for ReportCommentSvc<T> {
                        type Response = super::ReportCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ReportCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::report_comment(&inner, request).await
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
                        let method = ReportCommentSvc(inner);
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
                "/comment.api.v1.CommentService/PinComment" => {
                    #[allow(non_camel_case_types)]
                    struct PinCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PinCommentRequest>
                    for PinCommentSvc<T> {
                        type Response = super::PinCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PinCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::pin_comment(&inner, request).await
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
                        let method = PinCommentSvc(inner);
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
                "/comment.api.v1.CommentService/PageGetComment" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PageGetCommentRequest>
                    for PageGetCommentSvc<T> {
                        type Response = super::PageGetCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::page_get_comment(&inner, request)
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
                        let method = PageGetCommentSvc(inner);
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
                "/comment.api.v1.CommentService/PageGetSubComment" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetSubCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PageGetSubCommentRequest>
                    for PageGetSubCommentSvc<T> {
                        type Response = super::PageGetSubCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetSubCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::page_get_sub_comment(&inner, request)
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
                        let method = PageGetSubCommentSvc(inner);
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
                "/comment.api.v1.CommentService/PageGetSubCommentV2" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetSubCommentV2Svc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PageGetSubCommentV2Request>
                    for PageGetSubCommentV2Svc<T> {
                        type Response = super::PageGetSubCommentV2Response;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetSubCommentV2Request>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::page_get_sub_comment_v2(
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
                        let method = PageGetSubCommentV2Svc(inner);
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
                "/comment.api.v1.CommentService/PageGetDetailedComment" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetDetailedCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PageGetDetailedCommentRequest>
                    for PageGetDetailedCommentSvc<T> {
                        type Response = super::PageGetDetailedCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::PageGetDetailedCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::page_get_detailed_comment(
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
                        let method = PageGetDetailedCommentSvc(inner);
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
                "/comment.api.v1.CommentService/PageGetDetailedCommentV2" => {
                    #[allow(non_camel_case_types)]
                    struct PageGetDetailedCommentV2Svc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::PageGetDetailedCommentV2Request>
                    for PageGetDetailedCommentV2Svc<T> {
                        type Response = super::PageGetDetailedCommentV2Response;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::PageGetDetailedCommentV2Request,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::page_get_detailed_comment_v2(
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
                        let method = PageGetDetailedCommentV2Svc(inner);
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
                "/comment.api.v1.CommentService/GetPinnedComment" => {
                    #[allow(non_camel_case_types)]
                    struct GetPinnedCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::GetPinnedCommentRequest>
                    for GetPinnedCommentSvc<T> {
                        type Response = super::GetPinnedCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetPinnedCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::get_pinned_comment(&inner, request)
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
                        let method = GetPinnedCommentSvc(inner);
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
                "/comment.api.v1.CommentService/CountComment" => {
                    #[allow(non_camel_case_types)]
                    struct CountCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::CountCommentRequest>
                    for CountCommentSvc<T> {
                        type Response = super::CountCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CountCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::count_comment(&inner, request).await
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
                        let method = CountCommentSvc(inner);
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
                "/comment.api.v1.CommentService/BatchCountComment" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCountCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::BatchCountCommentRequest>
                    for BatchCountCommentSvc<T> {
                        type Response = super::BatchCountCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchCountCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::batch_count_comment(&inner, request)
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
                        let method = BatchCountCommentSvc(inner);
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
                "/comment.api.v1.CommentService/GetCommentLikeCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommentLikeCountSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::GetCommentLikeCountRequest>
                    for GetCommentLikeCountSvc<T> {
                        type Response = super::GetCommentLikeCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommentLikeCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::get_comment_like_count(
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
                        let method = GetCommentLikeCountSvc(inner);
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
                "/comment.api.v1.CommentService/GetCommentDislikeCount" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommentDislikeCountSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::GetCommentDislikeCountRequest>
                    for GetCommentDislikeCountSvc<T> {
                        type Response = super::GetCommentDislikeCountResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommentDislikeCountRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::get_comment_dislike_count(
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
                        let method = GetCommentDislikeCountSvc(inner);
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
                "/comment.api.v1.CommentService/CheckUserOnObject" => {
                    #[allow(non_camel_case_types)]
                    struct CheckUserOnObjectSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::CheckUserOnObjectRequest>
                    for CheckUserOnObjectSvc<T> {
                        type Response = super::CheckUserOnObjectResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckUserOnObjectRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::check_user_on_object(&inner, request)
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
                        let method = CheckUserOnObjectSvc(inner);
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
                "/comment.api.v1.CommentService/BatchCheckUserOnObject" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckUserOnObjectSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::BatchCheckUserOnObjectRequest>
                    for BatchCheckUserOnObjectSvc<T> {
                        type Response = super::BatchCheckUserOnObjectResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchCheckUserOnObjectRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::batch_check_user_on_object(
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
                        let method = BatchCheckUserOnObjectSvc(inner);
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
                "/comment.api.v1.CommentService/BatchCheckUserLikeComment" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckUserLikeCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<
                        super::BatchCheckUserLikeCommentRequest,
                    > for BatchCheckUserLikeCommentSvc<T> {
                        type Response = super::BatchCheckUserLikeCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::BatchCheckUserLikeCommentRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::batch_check_user_like_comment(
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
                        let method = BatchCheckUserLikeCommentSvc(inner);
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
                "/comment.api.v1.CommentService/BatchCheckCommentExist" => {
                    #[allow(non_camel_case_types)]
                    struct BatchCheckCommentExistSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::BatchCheckCommentExistRequest>
                    for BatchCheckCommentExistSvc<T> {
                        type Response = super::BatchCheckCommentExistResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::BatchCheckCommentExistRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::batch_check_comment_exist(
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
                        let method = BatchCheckCommentExistSvc(inner);
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
                "/comment.api.v1.CommentService/GetComment" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommentSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::GetCommentRequest>
                    for GetCommentSvc<T> {
                        type Response = super::GetCommentResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommentRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::get_comment(&inner, request).await
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
                        let method = GetCommentSvc(inner);
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
                "/comment.api.v1.CommentService/GetCommentUser" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommentUserSvc<T: CommentService>(pub Arc<T>);
                    impl<
                        T: CommentService,
                    > tonic::server::UnaryService<super::GetCommentUserRequest>
                    for GetCommentUserSvc<T> {
                        type Response = super::GetCommentUserResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommentUserRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as CommentService>::get_comment_user(&inner, request)
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
                        let method = GetCommentUserSvc(inner);
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
    impl<T> Clone for CommentServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "comment.api.v1.CommentService";
    impl<T> tonic::server::NamedService for CommentServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
