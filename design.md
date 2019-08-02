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


### Redux

Redux ha sido utilizado extensamente para manejar el estado de la aplicación. En mi opinion personal, una de las ventajas más grandes de `redux` es que le da uniformidad al manejo del estado, y permite centralizarlo sin necesidad de estar pasando todos los datos como propiedades.

Hay desarrolladores que dicen que con la llegada de los <a target="_blank" rel="noopener noreferrer" href="https://reactjs.org/docs/hooks-reference.html ">hooks</a>, ya librerías para manejar el estado como `redux` no serían necesarias. Mi opinión es que, al contrario, hace el ecosistema mucho más fuerte. Usar ambos le permite al desarrollador, por ejemplo, separar el estado del interfaz de usuario de un componente, que solo le corresponde al mismo, del estado global, donde pueden mantenerse los datos provenientes de una base de datos. Este es sólo uno de múltiples escenarios donde los `hooks` le dan libertades al desarrollador para separar las partes de la aplicación.

#### Re-ducks

El lector habrá notado que la implementación de `redux` no es la convencional. En el proyecto se sigue la propuesta descrita en <a target="_blank" rel="noopener noreferrer" href="https://github.com/erikras/ducks-modular-redux">https://github.com/erikras/ducks-modular-redux</a>, esencialmente por las mismas razones: es más fácil y más manejable mantener todos los elementos del flujo de datos en un solo archivo.

Con esta filosofía se pueden ir agregando secciones y modulos para el manejo del estado, y en caso de el archivo crecer demasiado, se puede separar en varios archivos distintos, siguiendo la separación de intereses.
