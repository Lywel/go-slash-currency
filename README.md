# Slash currency
This project is a currency implemented in Golang from scratch. It is secured by
an IBFT consensus algorithm. You can read more about it in its
[whitepaper](todo://link)

---
### Build the container
Some depandencies of this go-module are hosted on private repositories. You
*must* to be a collaborator in order to build it. Use one of your bitbucket
registered SSH keys as an argument in order to let Docker build the container
for you. This is a multi-step docker build, your private key is exclusively used
by the builder container and will *not* be included in the final one.

#### Add SSH Key in BitBucket

```sh
# Add the key to the ssh-agent if you don't want to type password each time you use the key
# ssh-add -K <private_key_path>
ssh-add -K ~/.ssh/id_rsa

# Copy your ssh public key to clipboard
# If you are on MacOS
# pbcopy < <private_key_path>.pub
cat ~/.ssh/id_rsa.pub
```

Then go to your [Account Setting](https://bitbucket.org/account) > SSH keys >
Add key. And paste the public key you copied previously. Now you can use ssh
to clone repositories without typing a password.

[Read more](https://confluence.atlassian.com/bitbucketserver/using-ssh-keys-to-secure-git-operations-776639772.html) about ssh configuration on bitbucket.

#### Docker Build

```sh
docker build --no-cache --build-arg SSH_KEY="$(cat ~/.ssh/id_rsa)" -t slash/validator-node .
# >Sending build context to Docker daemon  349.2kB
# >Step 1/10 : FROM golang:latest as builder
# ...
# >Successfully tagged slash/validator-node:latest
docker images slash/validator-node
# >REPOSITORY    ... SIZE
# >slash/validator-node ... 6.48MB
```

You should now have the `slash/validator-node` image and thanks to golang self-containness
it's less than 10MB in size.

---
### Start an instance
The container exposes port 8080. You can start as much instances on the same
machine as you want, mapping diffrent host ports to the listening port of the
containers. You should pass the ip of the other instances you want to connect
to.

```sh
# Local IP
ip='192.168.2.176'
# Instance 1 (listening on host:3000)
sudo docker run --rm -d -p 3000:8080 slash/validator-node:latest
# Instance 2 (listening on host:3001 and connecting to instance 1)
sudo docker run --rm -d -p 3001:8080 slash/validator-node:latest "$ip:3000"
# Instance 3 (listening on host:3002 and connecting to instances 1 & 2)
sudo docker run --rm -d -p 3002:8080 slash/validator-node:latest "$ip:3000" "$ip:3001"
# Instance 4 (listening on host:3003 and connecting to instance 3)
sudo docker run --rm -d -p 3003:8080 slash/validator-node:latest "$ip:3002"
```

Once an instance is launched you follow the logs (meaningless for now if you're
not a contributor). Find its container-id using `docker ps` and run `docker
logs <container-id>`.

