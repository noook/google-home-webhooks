# Google Home Webhooks

## Context

I just bought a Google Home Mini, and was wondering what was possible to make with it.
The goal was mainly to trigger code behind voice, so I tried to trigger something that I already
do through CLI: turning on my computer remotely with a WakeOnLAN.

I also heard about [IFTTT](https://ifttt.com) that allows you to automate a lot of things,
and an interesting feature is the webhooks. They allow me to make HTTP requests with JSON body
to a given URL, when an event (Google Home action) occurs.

## Idea

Webhooks are not secure, and everybody with the URL can trigger it, so I had to find a way
to be authenticated so only IFTTT can make requests. So I created a command that generates
a [JWT](https://en.wikipedia.org/wiki/JSON_Web_Token) that also attaches data to it. So when
the webhook receives the request, I can verify that I am the creator of the token, and can also 
read what's inside it. In this example, I'm attaching the MAC address of my computer.

## Usage

Setup your environment variables either by exporting it in your shell or by creating a `.env` file.
You need to fill in the `JWT_SECRET` and `SERVER_PORT` variables.

You can build the command then run it with :
```sh
go build -o ifttt-wol

./ifttt-wol generate <your-mac-address-here>
# ey.................
```

Keep this token and configure your IFTTT action, then pass the token in the `POST` request JSON
payload:
```json
{
  "token": "<your-token-here>"
}
```

Then run the server with
```sh
./ifttt-wol server
```

Your server is ready to accept hooks !