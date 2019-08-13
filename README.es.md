# Bibilioteca

Biblioteca en línea.

La versión en inglés de este READMe se encuentra en el archivo [README.md](./README.md).

[(English version.)](./README.md)

## Tabla de contenidos

- [Version en vivo](#live-version)
- [Proceso de diseño](#design-process)
- [Tecnologías](#technologies)
- [Filosofía de pruebas](#testing-philosophy)
- [Lanzamiento](#deployment)

  - [Docker](#docker)

    - [Requerimientos para correr el proyecto con Docker](#requirements-docker)
    - [Correr el proyecto local con Docker](#run-locally-docker)

  - [Build de producción](#staging-server)

## Version en vivo<a name="live-version"></a>

Las version en vivo de la aplicación se encuentra en <a target="_blank" rel="noopener noreferrer" href="https://library.djangulo.com">https://library.djangulo.com</a>.

El API se encuentra en <a target="_blank" rel="noopener noreferrer" href="https://library-api.djangulo.com">https://library-api.djangulo.com</a>. Cuenta con un índice que le indicará las rutas y ejemplos de las mismas. La versión en español está en <a target="_blank" rel="noopener noreferrer" href="https://library-api.djangulo.com/es">https://library-api.djangulo.com/es</a>

## Proceso de diseño<a name="design-process"></a>

Las decisiones tomadas y las razones detrás de las mismas están delineadas en el archivo [DESIGN.es.md](./DESIGN.es.md).

## Tecnologías utiizadas<a name="technologies"></a>

Las tecnologías centrales al desarrollo del cliente están listadas debajo. Nótese que esta no es una lista extensiva de todas las librerías usadas.

- Cliente

  - <a target="_blank" rel="noopener noreferrer" href="https://reactjs.org/">React</a>. Librería de JavaScript para construir interfaces de usuario. Inicializado a traves de <a target="_blank" rel="noopener noreferrer" href="https://github.com/facebook/create-react-app">CRA</a>.
  - <a target="_blank" rel="noopener noreferrer" href="https://redux.js.org/">Redux</a>. Para manejar el estado de la aplicación de manera centralizada.

- API

  - <a target="_blank" rel="noopener noreferrer" href="https://expressjs.com/">Express</a>. Una librería para construir aplicaciones de nodejs
  - La base de datos es <a target="_blank" rel="noopener noreferrer" href="https://www.postgresql.org/">PostgreSQL</a>.

- Lanzamiento

  - <a target="_blank" rel="noopener noreferrer" href="https://www.docker.com/">Docker</a>. Creación y manejo de contenedores.
  - <a target="_blank" rel="noopener noreferrer" href="https://traefik.io">Traefik</a>. Un reverse-proxy que confiere `TLS` por defecto. En este caso funciona como un <a target="_blank" rel="noopener noreferrer" href="https://es.wikipedia.org/wiki/Balance_de_carga">balanceador de carga</a>.
  - <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/">Amazon Web Services (AWS)</a>. Manejo de infraestructura.

## Filosofía de pruebas<a name="testing-philosophy"></a>

Es mayor enfoque de las pruebas automatizadas esta en integración y de punta a punta (E2E).
Las pruebas unitarias se utilizan estrictamente para funciones y utilidades, (validación, parsers, etc.), debido a que encuentro poco valor en las mismas. Esto se debe a que ya las pruebas de integración y e2e prueban el interfaz de usuario y el API a cabalidad.

En el directorio llamado `e2e/cypress/integration` se encuentran las diferentes pruebas de punta a punta.

Las pruebas de integración se encuentran al lado de cada archivo, bajo el formato de `*.spec.js`. Las pocas pruebas unitarias que se encuentren, también estarán bajo este formato.

## Lanzamiento<a name="deployment"></a>

### Docker

La manera más fácil y rápida de correr el proyecto de manera local, es a través de `Docker`.

#### Requerimientos para correr el proyecto con docker<a name="requirements-docker"></a>

- `docker`. <a target="_blank" rel="noopener noreferrer" href="https://docs.docker.com/install/linux/docker-ce/ubuntu/">Instrucciones de instalación</a>.
- `docker-compose`. <a target="_blank" rel="noopener noreferrer" href="(https://docs.docker.com/compose/install/">Instrucciones de instalación</a>.

#### Correr el proyecto local con Docker<a name="run-locally-docker"></a>

Lo primero es construir el contenedor:

```bash
~$ docker-compose -f local.yml build
```

Luego inicializarlo

```bash
~$ docker-compose -f local.yml up
```

Tras esto el proyecto puede visualizarse en `localhost:3000`

### Servidor de ensayo (staging server)<a name="staging-server"></a>

El servidor de ensayo (que será tambien utilizado como plataforma de prueba) se encuentra en <a target="_blank" rel="noopener noreferrer" href="https://library.djangulo.com">https://library.djangulo.com</a>.

Si desea lanzar su propio servidor, utilice el script `compose/production/traefik/aws_ec2_load_balancer`. Este script asume que tiene un dominio registrado con `AWS Route53`. Visite <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/domain-register.html">esta página</a> para ver instrucciones de como hacerlo.

Para que el script funcione, necesita contar con:

- <a target="_blank" rel="noopener noreferrer" href="https://aws.amazon.com/cli/">aws-cli</a>. <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html">Instalado</a> y <a target="_blank" rel="noopener noreferrer" href="https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html">configurado</a>.
- `docker`
- `docker-compose`
- `docker-machine`

El script configura una instancia `EC2`, utilizando `traefik` como un balanceador de carga, asumiendo que usted tiene el dominio `example.com` registrado en `AWS Route53`, lo utiliza de la siguiente manera:

```bash
~$ ./aws_ec2_load_balancer \
~$  --instance-name mi-instancia-ec2 \ # el nombre que EC2 le asigna a su instancia, es también el nombre por el cual docker-machine se refiere a la misma
~$  --open-ports 80,442 \ # puertos para abrir en la instancia
~$  --region us-east-1 \ # region de AWS en la cual crear la instancia
~$  --domain example.com \ #su dominio
~$  --subdomains api,library \ #registra api.example.com y library.example.com
~$  --networks library_api \ # networks para registrar las instancias de docker en el docker-compose
```

El script tomara unos minutos en terminar, una vez listo, tan solo debe de correr los comandos siguiente:

```bash
~$ eval $(docker-machine env mi-instancia-ec2) # "mi-instancia" es el nombre de su instancia EC2
~$ docker-compose -f production.yml build # se debe tomar unos minutos
~$ docker-compose -f production.yml up --detach
```

Si todo funcionó como se espera, su servidor debe de estar en línea.
