cd /home/tank/bsz/ryze/rpc/virtualNode-rpc

sudo docker build -t virtual-node-rpc:latest . &&  \
sudo docker tag virtual-node-rpc:latest bszpe/virtual-node-rpc:latest

cd /home/tank/bsz/ryze/rpc/dispatcher-rpc

sudo docker build -t dispatcher-rpc:latest . &&  \
sudo docker tag dispatcher-rpc:latest bszpe/dispatcher-rpc:latest

cd /home/tank/bsz/ryze/rpc/scheduler-rpc
sudo docker build -t scheduler-rpc:latest . &&  \
sudo docker tag scheduler-rpc:latest bszpe/scheduler-rpc:latest
