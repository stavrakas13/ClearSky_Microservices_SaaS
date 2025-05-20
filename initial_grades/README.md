# MICROSERVICE

# Grade Consumer README

This README covers how to set up MongoDB and RabbitMQ (with Docker), configure the environment for `app.js`, and test the grade‐import consumer with an Excel (`.xlsx`) file.

---

## 1. Prerequisites

* **Docker & Docker Compose** installed and running
* **Node.js** (v16+) locally if you run `app.js` outside of Docker
* **rabbitmqadmin** CLI (inside the RabbitMQ container or on host) for publishing test messages
* A sample Excel file (e.g. `test.xlsx`) containing the grade sheet

---

## 2. Environment Variables

Create a `.env` file in your project root with:

```dotenv
MONGO_URI=mongodb://mongo:27017/grades
RABBITMQ_URI=amqp://rabbitmq:5672
RABBITMQ_EXCHANGE=clearSky.event
RABBITMQ_ROUTING_KEY=postgrades.init
RABBIT_USER=yourRabbitUser
RABBIT_PASS=yourRabbitPass
```

* `MONGO_URI`: MongoDB connection string
* `RABBITMQ_URI`: RabbitMQ AMQP URI
* `RABBITMQ_EXCHANGE`: Exchange name your consumer listens on
* `RABBITMQ_ROUTING_KEY`: Routing key for grade‐upload messages
* `RABBIT_USER` / `RABBIT_PASS`: Management API credentials for testing

---

## 3. Docker Services for MongoDB & RabbitMQ

We’ll run **only** MongoDB and RabbitMQ in Docker. Your Node.js consumer (`app.js`) runs locally on your host.

### 3.1. Using `docker-compose`

Create a `docker-compose.yml` in your project root with:

```yaml
version: '3.8'

services:
  mongo:
    image: mongo:6
    restart: unless-stopped
    volumes:
      - mongo-data:/data/db

  rabbitmq:
    image: rabbitmq:3-management
    restart: unless-stopped
    ports:
      - '15672:15672' # Management UI
      - '5672:5672'   # AMQP
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBIT_USER}
      RABBITMQ_DEFAULT_PASS: ${RABBIT_PASS}

volumes:
  mongo-data:
```

Then:

```bash
# Start MongoDB and RabbitMQ
docker-compose up -d mongo rabbitmq
```

### 3.2. Or via individual `docker run` commands

If you don’t want `docker-compose`, you can run each service directly:

```bash
# MongoDB
docker run -d --name grades-mongo \
  -v mongo-data:/data/db \
  -p 27017:27017 \
  mongo:6

# RabbitMQ
docker run -d --name grades-rabbitmq \
  -p 15672:15672 -p 5672:5672 \
  -e RABBITMQ_DEFAULT_USER=${RABBIT_USER} \
  -e RABBITMQ_DEFAULT_PASS=${RABBIT_PASS} \
  rabbitmq:3-management
```

Once both are up, the management UI is at `http://localhost:15672`, and MongoDB listens on `localhost:27017`.

## 4. Running the Consumer Locally

Running the Consumer Locally

If you prefer to run `app.js` on your host machine:

```bash
# install deps
npm install

# ensure .env is in project root
node app.js
```

You should see:

```
✅ Connected to MongoDB database: grades
🚀 Waiting for messages on clearSky.event with routing key postgrades.init
```

---

## 5. Testing with RabbitMQ

1. **Copy your Excel into the RabbitMQ container**

   ```bash
   docker cp /home/stavros/Downloads/data/test.xlsx <rabbit_ctnr>:/tmp/test.xlsx
   ```

   Replace `<rabbit_ctnr>` with the RabbitMQ container name or ID.

2. **Publish a Base64 payload**

   ```bash
   docker exec -i <rabbit_ctnr> bash -lc '\
     rabbitmqadmin --vhost=/ \
       --username ${RABBIT_USER} --password ${RABBIT_PASS} \
       publish \
         exchange=${RABBITMQ_EXCHANGE} \
         routing_key=${RABBITMQ_ROUTING_KEY} \
         payload="$(base64 -w0 /tmp/test.xlsx)"'
   ```

3. **Observe your consumer logs**

   In the terminal running `node app.js` (or via `docker-compose logs -f grades-consumer`):

   ```
   📥 Received message
   ✅ Inserted 102 records into MongoDB
   ```

---

## 6. Verifying in MongoDB

Open a Mongo shell against the `grades` database:

```bash
mongo mongodb://localhost:27017/grades --eval "db.grades.find().pretty()"
```

You should see your imported documents.

---

## 7. Troubleshooting

* **Missing fields / validation errors**: Ensure your Excel headers match the mapping in `app.js` (e.g. `Κλίμακα βαθμολόγησης`).
* **Access refused**: Confirm `rabbitmqadmin` user has permissions (`administrator` tag + `set_permissions`).
* **Wrong database**: Check the `MONGO_URI` host\:path and verify in shell you’re using the same DB name.

---

Now you have a fully working Docker‐based setup for your grade consumer and a reproducible way to test with RabbitMQ and an XLSX file. Happy coding!
