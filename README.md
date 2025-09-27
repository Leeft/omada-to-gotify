# omada-to-gotify

## Purpose

This is a small program written in Go which spawns a server that'll receive
webhook messages from a TP-Link Omada Network Controller, it converts them
into Gotify notifications and delivers them to Gotify.

Run it in Docker, in a LXC, or really anywhere you like (anywhere as long as
the Omada Network Controller can talk to it, and it can talk to your Gotify
server). I'm running it together with Gotify in a docker-compose stack,
with the stack managed through Portainer. My compose file is further below.

![Gotify screenshot with example message](gotify.png)

## Installation / Configuration

Environment variables are used for configuration. They are:

### Required environment variables

- `GOTIFY_URL` - The base URL of your Gotify server (e.g., `https://gotify.example.com`). If you're using docker-compose in a stack with Gotify, just point to that directly (e.g. `http://gotify:80/`).
- `GOTIFY_APP_TOKEN` - The token for your Gotify application as configured inside Gotify.
- `OMADA_SHARED_SECRET` - The shared secret configured on the Omada Network Controller for this webhook.

### Optional environment variables

- `PORT` - The port on which to run the server (default is `8080`)

## Usage

To use this project directly without Docker:

1. Configure the webhook in Omada using the "Omada format", match the server and port where you are running this program. For example: `http://192.168.12.34:8080/`.
2. Set the required environment variables, making sure to include the shared secret from Omada.
3. Launch the executable with those environment variables set.
4. Enable the events to monitor in both the global view and your sites.
5. Wait for a message to come through from your Omada Controller and see it appear in Gotify.

At the moment there are no delivery retries should delivery fail, but each time it fails to either parse or deliver it will log an error to the console and then try connecting to Gotify again on the next request. However, Omada itself allows you to set up retries and see information about both successful and failed webhook requests so that should be adequate.

### docker

I've published a miniscule docker image `shiari/omada-to-gotify` at [Docker Hub](https://hub.docker.com/r/shiari/omada-to-gotify).

### docker-compose

I'm using the docker image in a stack together with Gotify. Here's my docker-compose.yml file as I use it in Portainer:

```yaml
services:

  gotify:
    image: gotify/server:latest
    ports:
      - "9400:80"
    volumes:
      - gotify_data:/app/data
    environment:
      GOTIFY_SERVER_CORS_ALLOWORIGINS: ${GOTIFY_SERVER_CORS_ALLOWORIGINS}
      GOTIFY_SERVER_STREAM_ALLOWEDORIGINS: ${GOTIFY_SERVER_STREAM_ALLOWEDORIGINS}
      GOTIFY_SERVER_TRUSTEDPROXIES: ${GOTIFY_SERVER_TRUSTEDPROXIES}
    restart: always

  omada-to-gotify:
    image: shiari/omada-to-gotify:latest
    environment:
      GOTIFY_URL: http://gotify:80/
      GOTIFY_APP_TOKEN: ${GOTIFY_APP_TOKEN}
      OMADA_SHARED_SECRET: ${OMADA_SHARED_SECRET}
    ports:
      - "8080:8080"

volumes:
  gotify_data:
    labels:
      - "com.example.gotify.description=Persistent volume for the gotify server"
```

Note that this will still need you to set up the environment variables in Portainer for both Gotify and this webhook proxy.

## Future

Possible additions to come (and feel free to contribute).

- Improving the instructions further, maybe also provide a basic LXC setup script.
- Specific support for more types of events from the Omada Controller, such as setting a different priority based on message contents or doing more to augment the information given. Some initial work is done towards this now though.
- Automated tests. Right now there aren't _any_ tests.
- MacOS support? I've got no way to test it works on MacOS, but I'll take pull requests for it if someone needs that. Then we'll blame you for any problems from then on. :wink:

## LICENSE

Copyright (c) 2025 Lianna Eeftinck <liannaee@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
