# Kubernetes Admission Control

```bash
kubectl create namespace example
kubectl config set-context --current --namespace example
make gencert
make deploy
make example
make clean
kubectl delete namespace example
```

## Resources

- <https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/>
