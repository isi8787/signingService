# Introduction 
TODO: Give a short introduction of your project. Let this section explain the objectives or the motivation behind this project. 

# Getting Started
TODO: Guide users through getting your code up and running on their own system. In this section you can talk about:
1.	Installation process
2.	Software dependencies
3.	Latest releases
4.	API references

# Build and Test
TODO: Describe and show how to build your code and run the tests. 

# Contribute
TODO: Explain how other users and developers can contribute to make your code better. 

If you want to learn more about creating good readme files then refer the following [guidelines](https://docs.microsoft.com/en-us/azure/devops/repos/git/create-a-readme?view=azure-devops). You can also seek inspiration from the below readme files:
- [ASP.NET Core](https://github.com/aspnet/Home)
- [Visual Studio Code](https://github.com/Microsoft/vscode)
- [Chakra Core](https://github.com/Microsoft/ChakraCore)

# Some Notes

- We only consider 2-out-of-3 threshold here.
  - signer 1 is the custody service.
  - signer 2 is the escrow service, it will NOT participate the signature generation,  but only participate key recovery.
  - signer 3 is the mobile client.
- WS stands for web socket.

--------

- To control who is joining the signature generation procedure, the argument pubSharesMap can be modified, which is a list of public shares that are related to the shares of the ECDSA private key.
  
`p.PrepareToSign(pk, k256Verifier, k256, proofParams, pubSharesMap, pubKeysMap)`

--------

## A potential _user package_

The package can include the following:

- The data structure of _user_, which include
  - User ID used in AD
  - User login credential
  - Common wallet operation supporting information
    - Paillier key pair
    - ZK common parameter
    - etc.
  - List of wallets (empty at first)
- Function: new user registration
- Function: wallet adding
- Function: TX creation (prepare for calling corresponding wallet function)
- Function: Query of wallets



## Connecting to DocumentDB locally

#Contact Isaac for newkey.pem

Run the following command
ssh -i "newkey.pem" -L 27017:docdb-2023-01-09-21-00-26.cluster-caltlurownlm.us-east-2.docdb.amazonaws.com:27017  ubuntu@ec2-18-188-77-19.us-east-2.compute.amazonaws.com  -fN
