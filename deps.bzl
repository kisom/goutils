load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "com_github_akavel_rsrc",
        importpath = "github.com/akavel/rsrc",
        sum = "h1:zjWn7ukO9Kc5Q62DOJCcxGpXC18RawVtYAGdz2aLlfw=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_certifi_gocertifi",
        importpath = "github.com/certifi/gocertifi",
        sum = "h1:6/yVvBsKeAw05IUj4AzvrxaCnDjN4nUqKjW9+w5wixg=",
        version = "v0.0.0-20180118203423-deb3ae2ef261",
    )
    go_repository(
        name = "com_github_cloudflare_backoff",
        importpath = "github.com/cloudflare/backoff",
        sum = "h1:8d1CEOF1xldesKds5tRG3tExBsMOgWYownMHNCsev54=",
        version = "v0.0.0-20161212185259-647f3cdfc87a",
    )
    go_repository(
        name = "com_github_cloudflare_cfssl",
        importpath = "github.com/cloudflare/cfssl",
        sum = "h1:vFJDAvQgFSRbCn9zg8KpSrrEZrBAQ4KO5oNK7SXEyb0=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_cloudflare_go_metrics",
        importpath = "github.com/cloudflare/go-metrics",
        sum = "h1:/8sZyuGTAU2+fYv0Sz9lBcipqX0b7i4eUl8pSStk/4g=",
        version = "v0.0.0-20151117154305-6a9aea36fb41",
    )
    go_repository(
        name = "com_github_cloudflare_redoctober",
        importpath = "github.com/cloudflare/redoctober",
        sum = "h1:p0Q1GvgWtVf46XpMMibupKiE7aQxPYUIb+/jLTTK2kM=",
        version = "v0.0.0-20171127175943-746a508df14c",
    )
    go_repository(
        name = "com_github_creack_pty",
        importpath = "github.com/creack/pty",
        sum = "h1:uDmaGzcdjhF4i/plgjmEsriH11Y0o7RKapEf/LDaM3w=",
        version = "v1.1.9",
    )
    go_repository(
        name = "com_github_daaku_go_zipexe",
        importpath = "github.com/daaku/go.zipexe",
        sum = "h1:VSOgZtH418pH9L16hC/JrgSNJbbAL26pj7lmD1+CGdY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_geertjohan_go_incremental",
        importpath = "github.com/GeertJohan/go.incremental",
        sum = "h1:7AH+pY1XUgQE4Y1HcXYaMqAI0m9yrFqo/jt0CW30vsg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_geertjohan_go_rice",
        importpath = "github.com/GeertJohan/go.rice",
        sum = "h1:KkI6O9uMaQU3VEKaj01ulavtF7o1fWT7+pk/4voiMLQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_getsentry_raven_go",
        importpath = "github.com/getsentry/raven-go",
        sum = "h1:ELaJ1cjF2nEJeIlHXahGme22yG7TK+3jB6IGCq0Cdrc=",
        version = "v0.0.0-20180121060056-563b81fc02b7",
    )
    go_repository(
        name = "com_github_go_sql_driver_mysql",
        importpath = "github.com/go-sql-driver/mysql",
        sum = "h1:7LxgVwFb2hIQtMm87NdgAVfXjnt4OePseqT1tKx+opk=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        importpath = "github.com/golang/protobuf",
        sum = "h1:YF8+flBXS5eO826T4nzqPrxfhQThhXl0YzfuUPu4SBg=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_google_certificate_transparency_go",
        importpath = "github.com/google/certificate-transparency-go",
        sum = "h1:Yf1aXowfZ2nuboBsg7iYGLmwsOARdV86pfH3g95wXmE=",
        version = "v1.0.21",
    )
    go_repository(
        name = "com_github_hashicorp_go_syslog",
        importpath = "github.com/hashicorp/go-syslog",
        sum = "h1:KaodqZuhUoZereWVIYmpUgZysurB1kBLX2j0MwMrUAE=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_jessevdk_go_flags",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_jmhodges_clock",
        importpath = "github.com/jmhodges/clock",
        sum = "h1:dYTbLf4m0a5u0KLmPfB6mgxbcV7588bOCx79hxa5Sr4=",
        version = "v0.0.0-20160418191101-880ee4c33548",
    )
    go_repository(
        name = "com_github_jmoiron_sqlx",
        importpath = "github.com/jmoiron/sqlx",
        sum = "h1:41Ip0zITnmWNR/vHV+S4m+VoUivnWY5E4OJfLZjCJMA=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_kisielk_sqlstruct",
        importpath = "github.com/kisielk/sqlstruct",
        sum = "h1:o/c0aWEP/m6n61xlYW2QP4t9424qlJOsxugn5Zds2Rg=",
        version = "v0.0.0-20150923205031-648daed35d49",
    )
    go_repository(
        name = "com_github_kisom_goutils",
        importpath = "github.com/kisom/goutils",
        sum = "h1:z4HEOgAnFq+e1+O4QdVsyDPatJDu5Ei/7w7DRbYjsIA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:mweAR1A6xJ3oS2pRaGiHgQ4OO8tzTaLawm8vnODuwDk=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_kr_fs",
        importpath = "github.com/kr/fs",
        sum = "h1:Jskdu9ieNAYnjxsi0LbQp1ulIKZV1LAFgK1tWhpZgl8=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_pty",
        importpath = "github.com/kr/pty",
        sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_kylelemons_go_gypsy",
        importpath = "github.com/kylelemons/go-gypsy",
        sum = "h1:mkl3tvPHIuPaWsLtmHTybJeoVEW7cbePK73Ir8VtruA=",
        version = "v0.0.0-20160905020020-08cad365cd28",
    )
    go_repository(
        name = "com_github_lib_pq",
        importpath = "github.com/lib/pq",
        sum = "h1:/qkRGz8zljWiDcFvgpwUpwIAPu3r07TDvs3Rws+o/pU=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:jbhqpg7tQe4SupckyijYiy0mJJ/pRyHvXf7JdWK860o=",
        version = "v1.10.0",
    )
    go_repository(
        name = "com_github_mreiferson_go_httpclient",
        importpath = "github.com/mreiferson/go-httpclient",
        sum = "h1:oKIteTqeSpenyTrOVj5zkiyCaflLa8B+CD0324otT+o=",
        version = "v0.0.0-20160630210159-31f0106b4474",
    )
    go_repository(
        name = "com_github_nkovacs_streamquote",
        importpath = "github.com/nkovacs/streamquote",
        sum = "h1:E2B8qYyeSgv5MXpmzZXRNp8IAQ4vjxIjhpAf5hv/tAg=",
        version = "v0.0.0-20170412213628-49af9bddb229",
    )
    go_repository(
        name = "com_github_op_go_logging",
        importpath = "github.com/op/go-logging",
        sum = "h1:lDH9UUVJtmYCjyT0CI4q8xvlXPxeZ0gYCVvWbmPlp88=",
        version = "v0.0.0-20160315200505-970db520ece7",
    )
    go_repository(
        name = "com_github_pkg_errors",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_pkg_sftp",
        importpath = "github.com/pkg/sftp",
        sum = "h1:/f3b24xrDhkhddlaobPe2JgBqfdt+gC/NYl0QY9IOuI=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_sirupsen_logrus",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:hI/7Q+DtNZ2kINb6qt/lS+IyXnHQe9e90POfeewL/ME=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_stretchr_objx",
        importpath = "github.com/stretchr/objx",
        sum = "h1:4G4v2dO3VZwixGIRoQ5Lfboy6nUhCyYzaqnIAPPhYs4=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        importpath = "github.com/stretchr/testify",
        sum = "h1:hDPOHmpOpP40lSULcqw7IrRb/u7w6RpDC9399XyoNd0=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_github_valyala_bytebufferpool",
        importpath = "github.com/valyala/bytebufferpool",
        sum = "h1:GqA5TC/0021Y/b9FG4Oi9Mr3q7XYx6KllzawFIhcdPw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_valyala_fasttemplate",
        importpath = "github.com/valyala/fasttemplate",
        sum = "h1:tY9CJiPnMXf1ERmG2EyK7gNUd+c6RKGD0IfU8WdUSz8=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_weppos_publicsuffix_go",
        importpath = "github.com/weppos/publicsuffix-go",
        sum = "h1:0Tu1uzLBd1jPn4k6OnMmOPZH/l/9bj9kUOMMkoRs6Gg=",
        version = "v0.13.0",
    )
    go_repository(
        name = "com_github_ziutek_mymysql",
        importpath = "github.com/ziutek/mymysql",
        sum = "h1:GB0qdRGsTwQSBVYuVShFBKaXSnSnYYC2d9knnE1LHFs=",
        version = "v1.5.4",
    )
    go_repository(
        name = "com_github_zmap_rc2",
        importpath = "github.com/zmap/rc2",
        sum = "h1:kKCF7VX/wTmdg2ZjEaqlq99Bjsoiz7vH6sFniF/vI4M=",
        version = "v0.0.0-20131011165748-24b9757f5521",
    )
    go_repository(
        name = "com_github_zmap_zcertificate",
        importpath = "github.com/zmap/zcertificate",
        sum = "h1:17HHAgFKlLcZsDOjBOUrd5hDihb1ggf+1a5dTbkgkIY=",
        version = "v0.0.0-20180516150559-0e3d58b1bac4",
    )
    go_repository(
        name = "com_github_zmap_zcrypto",
        importpath = "github.com/zmap/zcrypto",
        sum = "h1:PIpcdSOg3pMdFJUBg5yR9xxcj5rm/SGAyaWT/wK6Kco=",
        version = "v0.0.0-20200911161511-43ff0ea04f21",
    )
    go_repository(
        name = "com_github_zmap_zlint_v2",
        importpath = "github.com/zmap/zlint/v2",
        sum = "h1:b2kI/ToXX16h2wjV2c6Da65eT6aTMtkLHKetXuM9EtI=",
        version = "v2.2.1",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:qIbj1fsPNlZgppZ+VLlY7N33q108Sa+fhmuc+sWQYwY=",
        version = "v1.0.0-20180628173108-788fd7840127",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
        version = "v2.4.0",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:dUUwHk2QECo/6vqA44rthZ8ie2QXMNeKRTHCNY2nXvo=",
        version = "v3.0.0-20200313102051-9f266ea9e77c",
    )
    go_repository(
        name = "org_bitbucket_liamstask_goose",
        importpath = "bitbucket.org/liamstask/goose",
        sum = "h1:bkb2NMGo3/Du52wvYj9Whth5KZfMV6d3O0Vbr3nz/UE=",
        version = "v0.0.0-20150115234039-8488cc47d90c",
    )
    go_repository(
        name = "org_golang_x_crypto",
        importpath = "golang.org/x/crypto",
        sum = "h1:Qwe1rC8PSniVfAFPFJeyUkB+zcysC3RgJBAGk7eqBEU=",
        version = "v0.0.0-20220314234659-1baeb1ce4c0b",
    )
    go_repository(
        name = "org_golang_x_lint",
        importpath = "golang.org/x/lint",
        sum = "h1:5hukYrvBGR8/eNkX5mdUezrA6JiaEZDtJb9Ei+1LlBs=",
        version = "v0.0.0-20190930215403-16217165b5de",
    )
    go_repository(
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:CIJ76btIcR3eFI5EgSo6k1qKw9KJexJuRLI9G7Hp5wE=",
        version = "v0.0.0-20211112202133-69e39bad7dc2",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:ntjMns5wyP/fN65tdBD4g8J5w8n015+iIIs9rtjXkY0=",
        version = "v0.0.0-20220412211240-33da011f77ad",
    )
    go_repository(
        name = "org_golang_x_term",
        importpath = "golang.org/x/term",
        sum = "h1:v+OssWQX+hTHEmOBgwxdZxK4zHq3yOs8F9J7mk0PY8E=",
        version = "v0.0.0-20201126162022-7de9c90e9dd1",
    )
    go_repository(
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:aRYxNxv6iGQlyVaZmk6ZgYEDa+Jg18DxebPSrd6bg1M=",
        version = "v0.3.6",
    )
    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:/e+gpKk9r3dJobndpTytxS2gOy6m5uvpg+ISQoEcusQ=",
        version = "v0.0.0-20190311212946-11955173bddd",
    )
