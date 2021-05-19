Build and serve a bundle:

```
opa build policy.rego
python -m http.server 8000
```

Run the customized OPA:

```
go run main.go run \
    --set plugins.ticker=null
    --set services.http.url=http://localhost:8000
    --set bundles.bundle.trigger=manual
    --set bundles.bundle.resource=bundle.tar.gz
    --set status.console=true
    --server
```
