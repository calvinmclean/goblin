# GRPC

This example demonstrates the issue where different dependency versions cause plugin errors.

Maybe it would be better to just use HTTP instead of GRPC to reduce this. I could also remove the `cli` dependency and I would just be left with `github.com/miekg/dns` and it's sub-dependencies.
