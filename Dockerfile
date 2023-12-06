FROM golang:1.18-alpine


RUN mkdir -p /CustodyService/app/
WORKDIR /CustodyService/app
COPY ./signingService .

RUN wget https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem

ENV CosmosDbConnectionString="mongodb://fincodbadmin:newpassword@docdb-2023-01-09-21-00-26.cluster-caltlurownlm.us-east-2.docdb.amazonaws.com:27017/?ssl=true&tlsInsecure=true&ssl_ca_certs=rds-combined-ca-bundle.pem&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false"

ENV MONGODB_DATABASE="signer1-db"
ENV MONGODB_COLLECTION="signer1db-collection"
ENV PARTICIPANTID="1"
ENV AZURE_CLIENT_ID="765cb1df-e786-475f-82d5-5d7ebf9813f2"
ENV AZURE_CLIENT_SECRET="vWZ8Q~KK4PphJJ-KVUFBvN16TFOU6CYt-7M8wb6H"
ENV AZURE_TENANT_ID="2dc7c48e-0569-4780-805b-400ec5d480a1"
ENV AZURE_KEYVAULT_URL="https://fincovault.vault.azure.net/"

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./signingService"]
