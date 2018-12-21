# Slash currency
This project is a currency implemented in Golang from scratch. It is secured by
an IBFT consensus algorithm. You can read more about it in its
[whitepaper](todo://link)

---
## Preflight check
Most depandencies of this go-module are hosted on private repositories. You
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


---
## Get started using docker

### Build the docker container

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



---
## Get started without docker
### Dependencies

Go modules dependancies system uses git in the background. In order to let it
clone the private modules that are needed, your local git has to be configured
to use ssh instead of https. You can force by editing your git configuration
using this command for example:
```sh
git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
```

We used the experimental modules of go 1.11 so you need to have at least go
v1.11.
```sh
go version
# go version go1.11.3 linux/amd64
```

### Build
The build part can take from less than second to over a minute as go will
clone any missing dependencies on the first build.
```sh
git clone git@bitbucket.org:ventureslash/go-slash-currency.git
cd go-slash-currency
go build
```
### Start a node
Nodes have mandatory environmenet variables and a few optional flags.
`VAL_PORT` and `EP_PORT` are mandatory and respectively define wich port the
node will expose for ibft messages and http blockchain synchronization. Here is
the list of available optional flags.
```
Usage of ./go-slash-currency:
  -bc string
    	blockchain storage path (defaut: './chaindata') (default "./chaindata")
  -no-discovery
    	disable dns peer discovery
  -s value
    	address of a state provider
  -v value
    	address of a validator
  -w string
    	wallet file path (default "./slash-currency.wallet")
  -verbose-blockchain
    	print blockchain info level logs
  -verbose-core
    	print core info level logs
  -verbose-currency
    	print currency info level logs
  -verbose-endpoint
    	print endpoint info level logs
  -verbose-manager
    	print manager info level logs
  -verbose-network
    	print gossipnet info level logs
```

Here are a few example of start commands for diffrent purposes:
```
# Start and automatically join the main network
VAL_PORT=8080 EP_PORT=3000 ./go-slash-currency

# Start with dns peer discovery disabled. You will be alone
VAL_PORT=8080 EP_PORT=3000 ./go-slash-currency -no-discovery

# Start and join the main network plus some custom nodes you know.
VAL_PORT=8080 EP_PORT=3000 ./go-slash-currency \
  -s example.com:3000 -v example.com:8080 \
  -s 87.123.45.52:3001 -v 87.123.45.52:8000

# You can specify where to store the blockchain by using the -bc flag. If it
doesn't exist it will be created.
VAL_PORT=8080 EP_PORT=3000 ./go-slash-currency -bc 'path/to/chaindata'

# You can use a custom wallet by specifying a wallet path. If it doesn't exist
it will be generated.
VAL_PORT=8080 EP_PORT=3000 ./go-slash-currency -w path/to/mysuper.wallet
```

