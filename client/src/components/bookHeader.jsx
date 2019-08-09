import React from 'react';
import debounce from 'lodash/debounce';

import {
  Button,
  Form,
  Header,
  Icon,
  Pagination,
  Input,
  Menu
} from 'semantic-ui-react';

const BookHeader = ({
  error,
  sortColumn,
  sortDirection,
  sortByColumn,
  paginationItems,
  selectPaginationItem,
  currentPaginationItem,
  fetchBooks,
  searchQuery,
  setSearchQuery,
  searchBooks,
  setSearchItems,
  location
}) => {
  const resolveIcon = column => {
    if (sortColumn !== column) return null;
    if (sortDirection === 'asc') return 'triangle up';
    if (sortDirection === 'desc') return 'triangle down';
  };

  const sortOptions = [
    { key: 'title', value: 'title', text: 'Title' },
    { key: 'author', value: 'author', text: 'Author' },
    {
      key: 'pub-year',
      value: 'pub_year',
      text: 'Pub. Date'
    },
    { key: 'count', value: 'page_count', text: 'Pages' }
  ];

  // const resolveUrl = key => {
  //   const searchParams = new URLSearchParams(location.search);
  //   const qParams = {};
  //   for (let pair of searchParams.entries()) {
  //     qParams[pair[0]] = pair[1];
  //   }
  //   if (qParams.sort) console.log(qParams);
  // };

  // const resolvePaginationType = key => {
  //   if (key === 1) return 'firstItem';
  //   if (key === parseInt(currentPaginationItem.pages, 10)) return 'lastItem';
  //   return 'pageItem';
  // }
  const search = debounce(searchBooks, 600);

  return (
    <div>
      <Header as="h3">
        <Header.Content>Listings</Header.Content>
      </Header>
      <Menu secondary vertical>
        <Menu.Item>
          <Input
            placeholder="Search..."
            value={searchQuery}
            icon={
              searchQuery.length ? (
                <Icon
                  name="close"
                  link
                  onClick={() => {
                    setSearchQuery('');
                    setSearchItems([]);
                  }}
                />
              ) : (
                <Icon name="search" />
              )
            }
            onChange={(e, d) => {
              setSearchQuery(d.value);
              search(searchQuery, sortColumn, sortDirection);
            }}
            fluid
          />
        </Menu.Item>
        <Menu.Item>
          {currentPaginationItem && currentPaginationItem.pages ? (
            <Pagination
              defaultActivePage={1}
              ellipsisItem={false}
              totalPages={currentPaginationItem.pages}
              onPageChange={(e, d) => fetchBooks(d.activePage)}
            />
          ) : null}
        </Menu.Item>
        <Menu.Item>
          <Form.Group>
            <Form.Field>
              <label>Sort by</label>
              <Button.Group>
                {sortOptions.map(o => (
                  <Button
                    compact
                    size="mini"
                    key={o.key}
                    // as={NavLink}
                    // to={resolveUrl(key)}
                    onClick={() => sortByColumn(o.value)}
                    icon={resolveIcon(o.value)}
                    content={o.text}
                  />
                ))}
              </Button.Group>
            </Form.Field>
          </Form.Group>
        </Menu.Item>
      </Menu>
      {error ? <span>{error.toString()}</span> : null}
    </div>
  );
};

export default BookHeader;
