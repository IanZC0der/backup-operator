# backup-operator

based on K8S CRD+Operator


## install 
```bash
make install
```

## build operator
```bash
make docker-build docker-push IMG=ianzhang0405/backup-operator:v1beta4
```

## deploy
```bash
make deploy IMG=ianzhang0405/backup-operator:v1beta4
```

