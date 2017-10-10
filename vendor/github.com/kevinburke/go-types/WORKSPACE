git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    tag = "0.5.4",
)
load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_repositories", "go_repository")

go_repository(
    name = "com_github_satori_go_uuid",
    importpath = "github.com/satori/go.uuid",
    commit = "5bf94b69c6b68ee1b541973bb8e1144db23a194b",
)

go_repository(
    name = "in_gopkg_mgo_v2",
    importpath = "gopkg.in/mgo.v2",
    commit = "3f83fa5005286a7fe593b055f0d7771a7dce4655",
)

go_repositories()
