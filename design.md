# Proceso de diseño

El documento debajo delinea las decisiones de diseño y arquitectura a través de la aplicación, al igual que las razones detrás de las mismas.

Este es un documento viviente, y será modificado varias veces a lo largo del desarrollo de la aplicación.

## Tabla de contenidos

- [Aplicación meta](#goal-app)
- [Cliente](#client)
  - [¿Por qué React](#why-react)
  - [Docker](#really-docker)

## Aplicación meta

La meta es preparar y configurar una aplicación cliente-servidor, la cual estará compuesta por tres partes principales::

- El cliente (front-end)
- El API
- La base de datos

Por supuesto, preparar todo esto a la vez es muy ambicioso. Por tanto, comenzaremos con la apilcación de muestra entregada por Create-React-App, y nos enfocaremos en:

- La infraestructura del proyecto
- Preparar el desarrollo local
- Integración contínua
- Servidor de ensayo

Con estos elementos, la idea es crear un "pipeline" de desarrollo, donde podamos seguir un ciclos completos de desarrollo.

Dicho lo anterior, trataremos de mantener la planificación a largo plazo al mínimo, enfocandonos sólo en una sección a la vez, buscando integrarla al resto de la aplicación sin inconvenientes.

Iniciando con el cliente (front-end).

## Cliente<a name="client"></a>

El cliente será una aplicación web, más especificamente, una [Aplicación de Página Única](https://es.wikipedia.org/wiki/Single-page_application) (SPA, por sus siglas en inglés), creada con React.

### ¿Por qué React?<a name="why-react"></a>

React es una librería con un rico ecosistema que ofrece extendibilidad, la misma puede ser tan simple o tan compleja como se desee.

No solo esta cumpliendo con los requisitos (es una librería, no un framework), sino que evita la alta configuración inicial que requieren en ocasión frameworks como Angular. Esto hace React ideal para mi preferencia de desarrollo: desarrollo manejable, con implementaciones simples, de manera iterativa.

### Docker<a name="really-docker"></a>

El lector se estará preguntando: "¿por qué `docker`? ¿No se trata de una simple aplicación web?", y estaría en lo correcto, ya que es obvio que en la fase inicial de este proyecto, basta con un simple `yarn run start` o `npm run start` y el cliente estará corriendo.

Sin embargo, a medida que este proyecto evolucione, se agregarán partes que no resultarán tan simples como un comando de `npm`: me refiero a la base de datos y el API.

`docker` nos permite, de una manera determinada e inambigüa, tener todas las partes de nuestro proyecto en línea, en segundos.
