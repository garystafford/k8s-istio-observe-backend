# Go-services Environment Variables

```bash
export LOG_LEVEL=debug

export PROXY_PORT=8000
export SRV_A_PORT=50051
export SRV_B_PORT=$SRV_A_PORT
export SRV_C_PORT=$SRV_A_PORT
export SRV_D_PORT=$SRV_A_PORT
export SRV_E_PORT=$SRV_A_PORT
export SRV_F_PORT=$SRV_A_PORT
export SRV_G_PORT=$SRV_A_PORT
export SRV_H_PORT=$SRV_A_PORT

export SRV_A_URL=service-a
export SRV_B_URL=service-b
export SRV_C_URL=service-c
export SRV_D_URL=service-d
export SRV_E_URL=service-e
export SRV_F_URL=service-f
export SRV_G_URL=service-g
export SRV_H_URL=service-h

export MONGO_CONN=mongodb://mongodb:27017/admin
export RABBITMQ_CONN=amqp://guest:guest@rabbitmq:5672/

export SRV_C_DB=$SRV_C_URL
export SRV_F_DB=$SRV_F_URL
export SRV_G_DB=$SRV_G_URL
export SRV_H_DB=$SRV_H_URL

export SRV_D_QUEUE=$SRV_D_URL
export SRV_F_QUEUE=$SRV_F_URL

env | sort
```