# Bibilioteca

Bienvenidos, esta es mi entrega para la aplicación a GBH.

El `README.md` original de este repositorio fue movido al archivo llamado [instructions.md](./instructions.md).

## Tabla de contenidos

- [Version en vivo](#live-version)
- [Proceso de diseño](#design-process)
- [Tecnologías](#technologies)
- [Filosofía de pruebas](#testing-philosophy)
- [Lanzamiento](#deployment)
  - [Docker](#docker)
    - [Requerimientos para correr el proyecto con Docker](#requirements-docker)
    - [Correr el proyecto local con Docker](#run-locally-docker)
  
  - [Servidor de ensayo (staging server)](#staging-server)

## Version en vivo<a name="live-version"></a>

Las version en vivo de la aplicación se encuentra en <a target="_blank" rel="noopener noreferrer" href="https://library-staging.djangulo.com">https://library-staging.djangulo.com</a>

## Proceso de diseño<a name="design-process"></a>

Las decisiones tomadas y las razones detrás de las mismas están delineadas en el archivo [design.md](./design.md).

## Tecnologías utiizadas<a name="technologies"></a>

- Cliente

  - <a target="_blank" rel="noopener noreferrer" href="https://reactjs.org/">React</a>. Librería de JavaScript para construir interfaces de usuario. Inicializado a traves de <a target="_blank" rel="noopener noreferrer" href="https://github.com/facebook/create-react-app">CRA</a>.

- Lanzamiento

  - <a target="_blank" rel="noopener noreferrer" href="https://www.docker.com/">Docker</a>. Creación y manejo de contenedores.

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
docker-compose -f local.yml up
```

Tras esto el proyecto puede visualizarse en `localhost:3000`

### Servidor de ensayo (staging server)<a name="staging-server"></a>

El servidor de ensayo (que será tambien utilizado como plataforma de prueba) se encuentra en <a target="_blank" rel="noopener noreferrer" href="https://library-staging.djangulo.com">https://library-staging.djangulo.com</a>.

<!-- TODO como se lanza el servidor de ensayo -->
