# webhook
Simple webhook deployment for servers



- Go to [api github meta](https://api.github.com/meta) to get a list of IPs you need to allow through firewall
- create a firewall rule on port 3000 of above ips, for example I made: `192.30.252.0/22, 185.199.108.0/22, 140.82.112.0/20, 143.55.64.0/20, 2a0a:a440::/29, 2606:50c0::/32, 20.201.28.148/32, 20.205.243.168/32, 20.87.245.6/32, 20.248.137.49/32, 20.207.73.85/32, 20.27.177.116/32, 20.200.245.245/32, 20.175.192.149/32, 20.233.83.146/32, 20.29.134.17/32, 20.199.39.228/32, 4.208.26.200/32`
- for github, make a webhook e.g. https://github.com/jamfesteq/web/settings/hooks/new
- change ip from 999.999.999.999 to your server ip
- http://999.999.999.999:3000/webhook
- content type: application/json
- secret: generate a unique token yourself, put it in webhook.conf and this file
- send me everything
- make sure you are running the webhook with settings updated, save, and check if you get a ping event in webhook's stdout