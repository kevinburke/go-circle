http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.9.0/rules_go-0.9.0.tar.gz",
    sha256 = "4d8d6244320dd751590f9100cf39fd7a4b75cd901e1f3ffdfd6f048328883695",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")

go_repository(
    name = "com_github_kevinburke_go_uuid",
    importpath = "github.com/kevinburke/go.uuid",
    commit = "24443c65ec63d9e040fd4cedf0f1048b5d3544f7",
)

go_repository(
    name = "in_gopkg_mgo_v2",
    importpath = "gopkg.in/mgo.v2",
    commit = "3f83fa5005286a7fe593b055f0d7771a7dce4655",
)

go_rules_dependencies()
go_register_toolchains()
