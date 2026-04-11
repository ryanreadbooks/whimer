// @generated
pub mod comment {
    pub mod api {
        #[cfg(feature = "comment-api-v1")]
        // @@protoc_insertion_point(attribute:comment.api.v1)
        pub mod v1 {
            include!("comment/api/v1/comment.api.v1.rs");
            // @@protoc_insertion_point(comment.api.v1)
        }
    }
}
pub mod conductor {
    pub mod api {
        pub mod namespace {
            #[cfg(feature = "conductor-api-namespace-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.namespace.v1)
            pub mod v1 {
                include!("conductor/api/namespace/v1/conductor.api.namespace.v1.rs");
                // @@protoc_insertion_point(conductor.api.namespace.v1)
            }
        }
        pub mod namespaceservice {
            #[cfg(feature = "conductor-api-namespaceservice-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.namespaceservice.v1)
            pub mod v1 {
                include!("conductor/api/namespaceservice/v1/conductor.api.namespaceservice.v1.rs");
                // @@protoc_insertion_point(conductor.api.namespaceservice.v1)
            }
        }
        pub mod task {
            #[cfg(feature = "conductor-api-task-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.task.v1)
            pub mod v1 {
                include!("conductor/api/task/v1/conductor.api.task.v1.rs");
                // @@protoc_insertion_point(conductor.api.task.v1)
            }
        }
        pub mod taskservice {
            #[cfg(feature = "conductor-api-taskservice-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.taskservice.v1)
            pub mod v1 {
                include!("conductor/api/taskservice/v1/conductor.api.taskservice.v1.rs");
                // @@protoc_insertion_point(conductor.api.taskservice.v1)
            }
        }
        pub mod worker {
            #[cfg(feature = "conductor-api-worker-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.worker.v1)
            pub mod v1 {
                include!("conductor/api/worker/v1/conductor.api.worker.v1.rs");
                // @@protoc_insertion_point(conductor.api.worker.v1)
            }
        }
        pub mod workerservice {
            #[cfg(feature = "conductor-api-workerservice-v1")]
            // @@protoc_insertion_point(attribute:conductor.api.workerservice.v1)
            pub mod v1 {
                include!("conductor/api/workerservice/v1/conductor.api.workerservice.v1.rs");
                // @@protoc_insertion_point(conductor.api.workerservice.v1)
            }
        }
    }
}
pub mod counter {
    pub mod api {
        #[cfg(feature = "counter-api-v1")]
        // @@protoc_insertion_point(attribute:counter.api.v1)
        pub mod v1 {
            include!("counter/api/v1/counter.api.v1.rs");
            // @@protoc_insertion_point(counter.api.v1)
        }
    }
}
pub mod ext {
    #[cfg(feature = "ext-options")]
    // @@protoc_insertion_point(attribute:ext.options)
    pub mod options {
        include!("ext/options/ext.options.rs");
        // @@protoc_insertion_point(ext.options)
    }
}
pub mod msger {
    pub mod api {
        #[cfg(feature = "msger-api-msg")]
        // @@protoc_insertion_point(attribute:msger.api.msg)
        pub mod msg {
            include!("msger/api/msg/msger.api.msg.rs");
            // @@protoc_insertion_point(msger.api.msg)
        }
        pub mod system {
            #[cfg(feature = "msger-api-system-v1")]
            // @@protoc_insertion_point(attribute:msger.api.system.v1)
            pub mod v1 {
                include!("msger/api/system/v1/msger.api.system.v1.rs");
                // @@protoc_insertion_point(msger.api.system.v1)
            }
        }
        pub mod userchat {
            #[cfg(feature = "msger-api-userchat-v1")]
            // @@protoc_insertion_point(attribute:msger.api.userchat.v1)
            pub mod v1 {
                include!("msger/api/userchat/v1/msger.api.userchat.v1.rs");
                // @@protoc_insertion_point(msger.api.userchat.v1)
            }
        }
    }
}
pub mod note {
    pub mod api {
        #[cfg(feature = "note-api-v1")]
        // @@protoc_insertion_point(attribute:note.api.v1)
        pub mod v1 {
            include!("note/api/v1/note.api.v1.rs");
            // @@protoc_insertion_point(note.api.v1)
        }
    }
}
pub mod passport {
    pub mod api {
        pub mod access {
            #[cfg(feature = "passport-api-access-v1")]
            // @@protoc_insertion_point(attribute:passport.api.access.v1)
            pub mod v1 {
                include!("passport/api/access/v1/passport.api.access.v1.rs");
                // @@protoc_insertion_point(passport.api.access.v1)
            }
        }
        pub mod user {
            #[cfg(feature = "passport-api-user-v1")]
            // @@protoc_insertion_point(attribute:passport.api.user.v1)
            pub mod v1 {
                include!("passport/api/user/v1/passport.api.user.v1.rs");
                // @@protoc_insertion_point(passport.api.user.v1)
            }
        }
    }
}
pub mod relation {
    pub mod api {
        #[cfg(feature = "relation-api-v1")]
        // @@protoc_insertion_point(attribute:relation.api.v1)
        pub mod v1 {
            include!("relation/api/v1/relation.api.v1.rs");
            // @@protoc_insertion_point(relation.api.v1)
        }
    }
}
pub mod search {
    pub mod api {
        #[cfg(feature = "search-api-v1")]
        // @@protoc_insertion_point(attribute:search.api.v1)
        pub mod v1 {
            include!("search/api/v1/search.api.v1.rs");
            // @@protoc_insertion_point(search.api.v1)
        }
    }
}
pub mod wslink {
    pub mod api {
        pub mod forward {
            #[cfg(feature = "wslink-api-forward-v1")]
            // @@protoc_insertion_point(attribute:wslink.api.forward.v1)
            pub mod v1 {
                include!("wslink/api/forward/v1/wslink.api.forward.v1.rs");
                // @@protoc_insertion_point(wslink.api.forward.v1)
            }
        }
        pub mod protocol {
            #[cfg(feature = "wslink-api-protocol-v1")]
            // @@protoc_insertion_point(attribute:wslink.api.protocol.v1)
            pub mod v1 {
                include!("wslink/api/protocol/v1/wslink.api.protocol.v1.rs");
                // @@protoc_insertion_point(wslink.api.protocol.v1)
            }
        }
        pub mod push {
            #[cfg(feature = "wslink-api-push-v1")]
            // @@protoc_insertion_point(attribute:wslink.api.push.v1)
            pub mod v1 {
                include!("wslink/api/push/v1/wslink.api.push.v1.rs");
                // @@protoc_insertion_point(wslink.api.push.v1)
            }
        }
    }
}