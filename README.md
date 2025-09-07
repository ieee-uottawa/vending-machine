# IEEEuO-Machine-Controller

IEEE uOttawa Vending Machine controller using Pi 4

# Context

## Network

The Versatile Electronic Nutrient Dispenser (VEND) is runs on the BrotherLaserPrinter(BLP) network, which we think is a subnet of the subnet of eduroam that SITE is on (this should be verified). We enabled port forwarding on the BLP router for port 22, which enables us to SSH into the VEND without needing to be on the BLP network. However this does not work on any outside network.

## Implementation

The VEND works by leveraging Square Webhooks and the Sqaure API (for now). We run an HTTP server that receives the Webhooks when a payment for an item on our Square Store is fulfilled, then uses the Square API to get the information on which item was purchased and checks its custom attributes which contain the slot in the VEND where the item is located. It then dispenses the item. For the Square Webhooks to be able to reach the VEND, we need to expose a public domain to the internet. We do this using NGROK, which creates a free temporary domain which forwards traffic to the local port that we choose.

### Known Issues

- Whenever NGROK is restarted, the domain changes. This needs to also be reflected in the Square Developer Dashboard's Webhooks configuration, or else it will send the Webhooks to nothing.
- Sometimes the SSH connection disconnects but the server and ngrok are still running so you can't re-run them to get their GUI back. Run this command to kill them `pkill -f "go run ."; pkill -f "ngrok http 8000"; lsof -ti :8000 | xargs -r kill -9`

# Testing

## With the Square API

1. Log into the Square Developer Console
2. Navigate to Sandbox test accounts
3. Go to Docs & Tools in the navbar and open API Explorer in another tab
4. Open Square Dashboard for Default Test Account
5. In the Dashboard, go to Items & Services in the sidebar, Item Library and create an item
6. Give the item a name and a price and **MOST IMPORTANTLY** in the Stock section there is a dropdown called VendingMachineSlot(Square) which you need to assign one of the slots to the item.
7. Save the item
8. In the API Explorer, choose the Orders API
9. Go to Create order
10. Fill in the Access token, generate an idempotency key, fill the location_id, quantity
11. In the line_items\[0\] section, enter a quantity, in the catalog_object_id dropdown choose the ITEM_VARIATION_REGULAR option and choose item_type as ITEM
12. Choose the state as OPEN
13. Run the request and copy the order_id
14. Choose the Payments API
15. Go to Create payment
16. Generate idempotency_key
17. Put the source_id as cnon:card-nonce-ok
18. Put accept_partial_authorization as False
19. In amount_money, set the amount to anything higher than the item cost and the currency to CAD
20. Set autocomplete to true
21. Set the buyer_email_address to your email address
22. Set the delay_duration to P1W3d
23. Set the order_id to the one you copied earlier
24. Run the request

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
