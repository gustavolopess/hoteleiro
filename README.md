# Hoteleiro 🛎️

Hoteleiro a chat bot designed to help me to keep track of [Airbnb acomodations hosted by me](https://www.airbnb.com/users/show/114082252). Through telegram, I can:
- Add a rent registry
- Add an expense (like cleanings, bills, taxes, etc.)
- Register a mortgage payment
- Register a mortgage advance payment

All those informations are stored in a Google sheets by default - but the code architecture is flexible enough to accept any kind of storage.


### How to run locally
Install go dependencies
```bash
$ go mod install
```

Run it
```bash
$ go run cmd/main.go
```

### Next steps
- [ ] Inform the platform used to make the rent when adding a rent (AirBnb, Booking, Instagram, etc.)
- [ ] Add an allow-list of authorized Telegram Users
