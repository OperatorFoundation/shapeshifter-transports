cd transports

#cd transports/Dust/v2; go test -v
#cd transports/meeklite/v2; go test -v
#cd transports/meekserver/v2; go test -v
cd obfs2/v2; go build; go test -v
cd ../../obfs4/v2; go build; go test -v
cd ../../Optimizer/v2; go build; go test -v
cd ../../Replicant/v2; go build; go test -v
cd ../../shadow/v2; go build; go test -v
#cd transports/Dust/v3; go test -v
#cd transports/meeklite/v3; go test -v
#cd transports/meekserver/v3; go test -v
cd obfs2/v3; go build; go test -v
cd ../../obfs4/v3; go build; go test -v
cd ../../Optimizer/v3; go build; go test -v
cd ../../Replicant/v3; go build; go test -v
cd ../../shadow/v3; go build; go test -v


