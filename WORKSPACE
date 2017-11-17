http_archive(
    name = "io_bazel_rules_go",
    sha256 = "91fca9cf860a1476abdc185a5f675b641b60d3acf0596679a27b580af60bf19c",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.7.0/rules_go-0.7.0.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")

go_rules_dependencies()

go_register_toolchains()

go_repository(
    name = "org_golang_x_crypto",
    commit = "9f005a07e0d31d45e6656d241bb5c0f2efd4bc94",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "com_github_BurntSushi_toml",
    commit = "a368813c5e648fee92e5f6c30e3944ff9d5e8895",
    importpath = "github.com/BurntSushi/toml",
)

go_repository(
    name = "com_github_kevinburke_rest",
    commit = "b2c053a6ab954961c506c953ba0e735aac5725c2",
    importpath = "github.com/kevinburke/rest",
)

go_repository(
    name = "com_github_kevinburke_go_types",
    commit = "5a1b614c70f6a01aa3415dbdd790752ae87a8da3",
    importpath = "github.com/kevinburke/go-types",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "fd80eb99c8f653c847d294a001bdf2a3a6f768f5",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "in_gopkg_mgo_v2",
    commit = "3f83fa5005286a7fe593b055f0d7771a7dce4655",
    importpath = "gopkg.in/mgo.v2",
)

go_repository(
    name = "com_github_satori_go_uuid",
    commit = "5bf94b69c6b68ee1b541973bb8e1144db23a194b",
    importpath = "github.com/satori/go.uuid",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "bf42f188b9bc6f2cf5b8ee5a912ef1aedd0eba4c",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "com_github_inconshreveable_log15",
    commit = "0decfc6c20d9ca0ad143b0e89dcaa20f810b4fb3",
    importpath = "github.com/inconshreveable/log15",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    commit = "6ca4dbf54d38eea1a992b3c722a76a5d1c4cb25c",
    importpath = "github.com/mattn/go-isatty",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    commit = "6fcc0c1fd9b620311d821b106a400b35dc95c497",
    importpath = "github.com/mattn/go-colorable",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    strip_prefix = "stack-54be5f394ed2c3e19dac9134a40a95ba5a017f7b",
    type = "zip",
    urls = ["https://codeload.github.com/go-stack/stack/zip/54be5f394ed2c3e19dac9134a40a95ba5a017f7b"],
)

go_repository(
    name = "com_github_kevinburke_go_git",
    commit = "da65f3cfe562df3928c9fd53b24030389ba00cd7",
    importpath = "github.com/kevinburke/go-git",
)

go_repository(
    name = "com_github_google_go_cmp",
    commit = "98232909528519e571b2e69fbe546b6ef35f5780",
    importpath = "github.com/google/go-cmp",
)

go_repository(
    name = "com_github_skratchdot_open_golang",
    commit = "75fb7ed4208cf72d323d7d02fd1a5964a7a9073c",
    importpath = "github.com/skratchdot/open-golang",
)

go_repository(
    name = "com_github_kevinburke_bigtext",
    commit = "fd72dcfb912b32532896bbc98413e5bd425eb315",
    importpath = "github.com/kevinburke/bigtext",
)
