# Tendermint RPC

## Generate markdown for [Slate](https://github.com/bcbchain/tendermint/slate)

We are using [Slate](https://github.com/bcbchain/tendermint/slate) to power our RPC
documentation. If you are changing a comment, make sure to copy the resulting
changes to the slate repo and make a PR
[there](https://github.com/bcbchain/tendermint/slate) as well. For generating markdown
use:

```shell
go get github.com/melekes/godoc2md

godoc2md -template rpc/core/doc_template.txt github.com/bcbchain/tendermint/rpc/core | grep -v -e "pipe.go" -e "routes.go" -e "dev.go" | sed 's$/src/target$https://github.com/bcbchain/tendermint/tree/master/rpc/core$'
```
