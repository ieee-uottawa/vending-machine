# IEEEuO-Machine-Controller

IEEE uOttawa Vending Machine controller using Pi 4

# Context

## Network

The Versatile Electronic Nutrient Dispenser (VEND) is runs on the BrotherLaserPrinter(BLP) network, which we think is a subnet of the subnet of eduroam that SITE is on (this should be verified). We enabled port forwarding on the BLP router for port 22, which enables us to SSH into the VEND without needing to be on the BLP network. However this does not work on any outside network.

## Implementation

The VEND works by leveraging Square Webhooks and the Sqaure API (for now). We run an HTTP server that receives the Webhooks when a payment for an item on our Square Store is fulfilled, then uses the Square API to get the information on which item was purchased and checks its custom attributes which contain the slot in the VEND where the item is located. It then dispenses the item. For the Square Webhooks to be able to reach the VEND, we need to expose a public domain to the internet. We do this using NGROK, which creates a free temporary domain which forwards traffic to the local port that we choose.

### Known Issues

- Whenever NGROK is restarted, the domain changes. This needs to also be reflected in the Square Developer Dashboard's Webhooks configuration, or else it will send the Webhooks to nothing.

# SSHing into the Pi

## From the BrotherLaserPrinter network

1. Open a terminal or preferably VS Code
2. ```bash
   ssh ieeepi@192.168.1.102
   ```
3. Enter the password

## From the eduroam network (in the office only for now, need to test elsewhere)

1. Open a terminal or preferably VS Code
2. ```bash
   ssh ieeepi@10.136.193.96
   ```
3. Enter the password

# Running the server

## Python Server

1. Open a terminal
2. ```bash
   cd /home/ieeepi/dev/vending-machine/python-server
   ```
3. ```bash
   python3 dispense.py
   ```
4. Open another terminal
5. ```bash
   ngrok http 8000
   ```

## Go Server

1. Open a terminal
2. ```bash
   cd /home/ieeepi/dev/vending-machine/go-server
   ```
3. ```bash
   go run .
   ```
4. Open another terminal
5. ```bash
   ngrok http 8000
   ```
