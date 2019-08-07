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

### Estructura del projecto

```bash
├── public              # archivos publicos del cliente
├── src
│   ├── components      # componentes de react
│   ├── data            # data para pruebas unitarias
│   ├── services        # interfaces con servicios externos
│   ├── store           # configuración de Redux
│   ├── App.css
│   ├── App.js
│   ├── config.js
│   ├── configureStore.js
│   ├── index.css
│   ├── index.js
│   ├── logo.svg
│   └── serviceWorker.js
├── package.json
└── yarn.lock
```

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

## API

Para la estructura del API, se utilizó `expressjs`, debido a su fácil utilización, y modularidad. La base de datos fue `PostgreSQL`; entiendo que no es la primera opción cuando se trata de aplicaciones de `node`, sin embargo, cuento con bastante experiencia trabajando con `postgres`: adaptar la librería `pg-promise` para utilizarla con `express` fue relativamente simple.

### Estructura del directorio

```bash
├── src
│   ├── db              # archivos relacionados a la base de datos
│   │   ├── migrations  # migraciones a la base de datos
│   ├── routes          # rutas del API
│   ├── utils           # utilidades
│   ├── config.js       # configuración del projecto
│   ├── index.js        # raíz del API
├── package.json
├── package-lock.json
└── yarn.lock
```

### Modelado de la base de datos

La base de datos fue modelada de la manera más simple posible, donde un libro cuenta con N objetos relacionados (llamados páginas).

Existe la posibilidad de crear otros modelos para darle más especificidad y estructura a la tabla `books` (libros), tales como parte, capítulo, sección, etc. Esta extensión, a pesar de ofrecer más estructura, trae consigo nuevos retos, particularmente en el procesamiento automático de los libros, debido a que hay que extraer los títulos, capítulos y secciones con una librería de procesamiento de lenguaje natural, como lo es `NLTK`. Esta tarea, aunque interesante, se sale del alcance de este proyecto.

En la fase inicial que nos encontramos, y en búsqueda del producto mínimo viable (MVP, por sus siglas en inglés), nos abstendrémos de extender los modelos más alla de libros y páginas.

### Alimentación de la base de datos

La base de datos es alimentada con un cuerpo de datos de <a target="_blank" rel="noopener noreferrer" href="https://www.nltk.org/">NLTK (Natural Language Toolkit)</a>, una librería de <a target="_blank" rel="noopener noreferrer" href="https://es.wikipedia.org/wiki/Procesamiento_de_lenguajes_naturales">procesamiento de lenguaje natural</a> de <a target="_blank" rel="noopener noreferrer" href="https://python.org">Python</a>. Este cuerpo de datos de `NLTK` es descargado por automáticamente al inicializar el projecto con `docker`. En caso de lanzar el proyecto de manera manual, el script `api/src/db/init_nltk.py`, instalará `nltk` y descargará el `gutenberg corpora` al directorio `api/src/db/nltk_data/`, con los siguientes archivos:


```bash
├── src
│   ├── db
│   │   ├── ...
│   │   ├── nltk_data
│   │   │   └── corpora
│   │   │       ├── gutenberg
│   │   │       │   ├── austen-emma.txt
│   │   │       │   ├── austen-persuasion.txt
│   │   │       │   ├── austen-sense.txt
│   │   │       │   ├── bible-kjv.txt
│   │   │       │   ├── blake-poems.txt
│   │   │       │   ├── bryant-stories.txt
│   │   │       │   ├── burgess-busterbrown.txt
│   │   │       │   ├── carroll-alice.txt
│   │   │       │   ├── chesterton-ball.txt
│   │   │       │   ├── chesterton-brown.txt
│   │   │       │   ├── chesterton-thursday.txt
│   │   │       │   ├── edgeworth-parents.txt
│   │   │       │   ├── melville-moby_dick.txt
│   │   │       │   ├── milton-paradise.txt
│   │   │       │   ├── README
│   │   │       │   ├── shakespeare-caesar.txt
│   │   │       │   ├── shakespeare-hamlet.txt
│   │   │       │   ├── shakespeare-macbeth.txt
│   │   │       │   └── whitman-leaves.txt
│   │   │       └── gutenberg.zip
│   │   ├── ...
```

Al inicializarse el servidor del API, la aplicación de `express` se conecta con la base de datos de `postgres`, y utilizando métodos asíncronos descritos en `api/src/db/seeddb.js`, se procesa cada archivo en `api/src/db/nltk_data/corpora/gutenberg/` de la siguiente manera:

1. Se extraen los datos del libro
2. Se divide cada libro en "párrafos" (efectivamente, un arreglo de párrafos)
3. Se asigna a cada página de un libro N párrafos (N está descrito en `api/src/config.js` bajo el nombre de `paragraphsPerPage`).
4. Se crea un arreglo para libros (`books`) y un arreglo para páginas (`pages`), que serán guardados en `api/src/db/seed_data/books.json` y `api/src/db/seed_data/pages.json`, respectivamente.
    - Estos archivos serán reutilizados cada vez que se reinicie el servidor.
5. La aplicación tratará de re-alimentar los datos en `api/src/db/seed_data/`, pero debido a una declaración `ON CONFLICT ...` de `postgres`, no duplicará contenidos.

La idea detrás de los archivos en `api/src/db/seed_data/` es crear una especie de cache, ya que la operación de leer los archivos de texto y procesarlos puede ser costosa.

Dichos archivos no pueden ser sometidos al control de versiones debido a que tienen alrededor de `13MB`.