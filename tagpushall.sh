#git tag $1
#git push origin $1

for transport in obfs4 obfs2 shadow Replicant Optimizer Dust meeklite meekserver StarBridge
do
  ./tagpush.sh $transport $1
done
