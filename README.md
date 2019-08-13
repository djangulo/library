# Library

Online library application.

The spanish version of this README its in the [README.es.md](./README.es.md) file
.

[(Versión en español.)](./README.es.md)

## Tabla de contenidos

- [Live version](#live-version)
- [Design process](#design-process)
- [Technologies](#technologies)
- [Testing philosophy](#testing-philosophy)
- [Deployment](#deployment)

  - [Docker](#docker)

    - [Docker launch requirements](#requirements-docker)
    - [Run locally with Docker](#run-locally-docker)

  - [Production build](#production-server)

## Live version<a name="live-version"></a>

Live version of the app is found at <a target="_blank" rel="noopener noreferrer" href="https://library.djangulo.com">https://library.djangulo.com</a>.

The API can be found at <a target="_blank" rel="noopener noreferrer" href="https://library-api.djangulo.com">https://library-api.djangulo.com</a>. It has an index that lists available routes and examples of them. The spanish version is at <a target="_blank" rel="noopener noreferrer" href="https://library-api.djangulo.com/es">https://library-api.djangulo.com/es</a>.

## Design process<a name="design-process"></a>

Design choices and the reasons behind them are outlined in the [DESIGN.md](./DESIGN.md) file.

## Technologies<a name="technologies"></a>

The central technologies to the creation of the client ar elisted below. Note thatthis is by no means an exhaustive list.

- Client

  - <a target="_blank" rel="noopener noreferrer" href="https://reactjs.org/">React</a>. JavaScript library to create user interfaces. Initialized through <a target="_blank" rel="noopener noreferrer" href="https://github.com/facebook/create-react-app">CRA</a>.
  - <a target="_blank" rel="noopener noreferrer" href="https://redux.js.org/">Redux</a>. State management.

- API

  - <a target="_blank" rel="noopener noreferrer" href="https://expressjs.com/">Express</a>. NodeJS application framework.
  - The database is <a target="_blank" rel="noopener noreferrer" href="https://www.postgresql.org/">PostgreSQL</a>.

- Deployment

  - <a target="_blank" rel="noopener noreferrer" href="https://www.docker.com/">Docker</a>. Containerization.
  - <a target="_blank" rel="noopener noreferrer" href="https://traefik.io">Traefik</a>. Reverse-proxy with automatic `TLS`. In this case is used as a <a target="_blank" rel="noopener noreferrer" href="https://en.wikipedia.org/wiki/Load_balancing_(computing)">load balancer.</a>.
  - <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/">Amazon Web Services (AWS)</a>. Infrastructure as a service.

## Testing philosophy<a name="testing-philosophy"></a>

The biggest focus of the tests is on integration and End to End (E2E). Unit tests are used sparsly in the client, as E2E and integration tests cover most, if not all cases.

The directory `e2e/cypress/integration` has the client E2E tests.

Integration tests follow the `*spec.js` convention.

## Deployment<a name="deployment"></a>

### Docker

The easiest way to get the project up and running locally is by using `Docker`.

#### Docker requirements<a name="requirements-docker"></a>

- `docker`. <a target="_blank" rel="noopener noreferrer" href="https://docs.docker.com/install/linux/docker-ce/ubuntu/">Installation instructions.</a>.
- `docker-compose`. <a target="_blank" rel="noopener noreferrer" href="(https://docs.docker.com/compose/install/">Installation instructions.</a>.

#### Run locally with `docker`<a name="run-locally-docker"></a>

First thing is to build the container:

```bash
~$ docker-compose -f local.yml build
```

Then initialize it:

```bash
~$ docker-compose -f local.yml up
```

It can be found at `localhost:3000`

### Production build<a name="production-server"></a>

The production server, which will also be used as a testing platform, is found at <a target="_blank" rel="noopener noreferrer" href="https://library.djangulo.com">https://library.djangulo.com</a>.

If you wish to launch your own, feel free to use the `compose/production/traefik/aws_ec2_load_balancer` script. This script assumes that you have an `AWS Route53` hosted zone with a custom domain name. Visit <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/domain-register.html">this page</a> to see instructions for registering one.

For the script to work, you will need:

- <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/cli/">aws-cli</a>. <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html">Installed</a> and <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html">configured</a>.
- [docker]("#requirements-docker")
- [docker-compose]("#requirements-docker")
- <a target="_blank" rel="noopener noreferrer" href="https://docs.docker.com/machine/install-machine/">docker-machine</a>

The script configures an `EC2` instance, using `traefik` as a load balancer. Assuming you have the `example.com` domain registered in `AWS Route53`, you can use it as follows:

```bash
~$ ./aws_ec2_load_balancer \
~$  --instance-name mi-instancia-ec2 \ # name EC2 assigns to the instance, it's also the name of the docker-machine
~$  --open-ports 80,442 \ # ports to open in the instance
~$  --region us-east-1 \ # AWS region to create the instance in
~$  --domain example.com \ # your domain
~$  --subdomains api,library \ # registers A records for api.example.com y library.example.com
~$  --networks library_api \ # networks to register the docker instances in the docker-compose
```

The script should take a few minutes to run. Once ready, you just need to run the following commands:

```bash
~$ eval $(docker-machine env my-instance) # "mf-instance" is the name you gave your EC2 instance
~$ docker-compose -f production.yml build # should take a few minutes
~$ docker-compose -f production.yml up --detach
```

If everything worked as expected, your server should be online.
