URL="http://localhost:8089/calls?callID=198d9723-8bcb-41b4-998b-8d04be2d068a&name=test&phone=01022876977"  # proxy 
URL_CAL="http://localhost:8089/calls?token=84c4eb27-1dc7-4c26-a0b2-03299e3ed61e&callID=198d9723-8bcb-41b4-998b-8d04be2d068a&name=test&phone=01022876977"
random_string=$(openssl rand -hex 6)

# Evaluate transactionId with the random_string
transactionId="198d9723-8bcb-41b4-998b-${random_string}"

echo $transactionId

time curl -k --location "$URL_CAL" \
--header 'Authorization: ApiKey a05a346016cd93fffe620a8f51e221a4' \
--header 'Content-Type: application/json' 
# --data '{
#          "name": "testvoice",
#          "transactionId": "'"$transactionId"'",  
#          "to": {
#            "subscriberId": "201002071244",
#             "phone": "201002071244"
#          }
#        }'
