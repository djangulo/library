const fs = require('fs');
const uuidv4 = require('uuid/v4');
const books = [
  {
    id: 'de0e4051-54b1-4f37-97f2-619b5b568d7f',
    title: 'Moby Dick',
    slug: 'moby-dick',
    author: 'Herman Melville',
    synopsis: 'A sailor is obsessed with a white whale.',
    date_added: '2019-07-27',
    publication_date: '1851-10-18',
    page_count: 20
  },
  {
    id: '9c3e08f9-2a5b-4eed-8769-991a708fa036',
    title: 'The old man and the sea',
    slug: 'the-old-man-and-the-sea',
    author: 'Ernest Hemingway',
    synopsis:
      "One of his [Hemingway's] most famous works, it tells the story of Santiago, an aging Cuban fisherman who struggles with a giant marlin far out in the Gulf Stream off the coast of Cuba.",
    date_added: '2019-07-27',
    publication_date: '1952',
    isbn: '0-684-80122-1',
    page_count: 20
  },
  {
    id: '62a01525-a3b1-49cb-919c-d83f0aa10d2f',
    title: 'Carrie',
    slug: 'carrie',
    author: 'Stephen King',
    synopsis:
      'Carrie is an epistolary horror novel by American author Stephen King.',
    date_added: '2019-07-27',
    publication_date: '1974-04-05',
    isbn: '978-0-385-08695-0',
    page_count: 20
  },
  {
    id: 'ae31188e-b343-4987-8842-c97afb5f0c6d',
    title: 'It',
    slug: 'it',
    author: 'Stephen King',
    synopsis:
      'It is a 1986 horror novel by American author Stephen King. It was his 22nd book, and his 18th novel written under his own name.',
    date_added: '2019-07-27',
    publication_date: '1986-09-15',
    isbn: '0-670-81302-8',
    page_count: 20
  },
  {
    id: '6cb80e34-0394-43ac-957d-7cefddb6e493',
    title: 'Pet Sematary',
    slug: 'pet-sematary',
    author: 'Stephen King',
    synopsis:
      'Pet Sematary is a 1983 horror novel by American writer Stephen King.',
    date_added: '2019-07-27',
    publication_date: '1983-11-14',
    isbn: '	978-0-385-18244-7',
    page_count: 20
  },
  {
    id: '8c79ac56-39f2-4954-8a1f-cd3b058c169f',
    title: 'The call of Cthulu',
    slug: 'the-call-of-cthulu',
    author: 'H.P. Lovecraft',
    synopsis:
      '"The Call of Cthulhu" is a short story by American writer H. P. Lovecraft. Written in the summer of 1926, it was first published in the pulp magazine Weird Tales, in February 1928.',
    date_added: '2019-07-27',
    publication_date: '1928-02',
    page_count: 20
  }
];

function create_pages(books) {
  const pages = [];
  for (let book of books) {
    for (let i = 1; i <= 20; i++) {
      pages.push({
        id: uuidv4(),
        book: book['id'],
        number: i,
        text: `${book.title} page ${i}`
      });
    }
  }
  return pages;
}

const pages = create_pages(books);
const jsonPages = JSON.stringify(pages, null, 2);
const jsonBooks = JSON.stringify(books, null, 2);

fs.writeFile('./fakePages.json', jsonPages, 'utf8', () =>
  console.log('fakePages.json written')
);

fs.writeFile('./fakeBooks.json', jsonBooks, 'utf8', () =>
  console.log('fakeBooks.json written')
);
