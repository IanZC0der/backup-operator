# backup-operator

based on K8S CRD+Operator


## install 
```bash
make install
```

## build operator
```bash
make docker-build docker-push IMG=harbor.kubecenter.com/operator/backup-operator:v1beta1
```

## deploy
```bash
make deploy IMG=harbor.kubecenter.com/operator/backup-operator:v1beta1
```

