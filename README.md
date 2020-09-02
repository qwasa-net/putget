# putget 

## putget — my little **go** experiment

**putget** is a super stupid http service to:

* upload series of files into buckets

* get the most recent file from the bucket


### Run *putget* service

```bash
> go run putget.go -bind "localhost:8900" -filesroot /tmp/putget.files/ -urlroot
```


### Use *putget*
```bash

> curl -X POST http://localhost:8900/bucket_name/ --data-binary @image1.jpg
> curl -X POST http://localhost:8900/bucket_name/ --data-binary @imageN.jpg
…
> curl http://localhost:8900/bucket_name/ -O downloaded_image.jpg

```

## etc

* Authorization, limits and access control must be handled in the router/proxy service, e.g. **nginx**.

* DataBase module is not implemented (yet?) — dumb memory map is used

* This is a toy project, not intended for production use.

:wq
