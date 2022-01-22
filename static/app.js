const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const data = Object.fromEntries(new FormData(form));
    const searchData = data.query;
    const searchParams = new URLSearchParams(window.location.search);
    const page = searchParams.get('page') || '1'
    Controller.searchQuery(searchData, page)
  },

  searchQuery: (query, page) => {
    const input = document.getElementById("query")
    input.value = query;
    const searchParams = new URLSearchParams(window.location.search);
    const response = fetch(`/search?q=${query}&limit=10&page=${page}`).then((response) => {
      response.json().then((res) => {
        Controller.renderResponse(res, query);
      });
    });
  },

  renderResponse: (res, searchData) => {
    const table = document.getElementById("table-body");
    const rows = [];
    for (let result of res.results) {
      let re = new RegExp(`(${searchData})`, "gi");
      console.info("REEE", re)
      result = result.replace(re, "<b>$1</b>")
      rows.push(`<tr>${result}<tr/>`);
    }
    table.innerHTML = rows;
    const pagination = document.getElementById("pagination");
    const pages = []
    for (let i = 1; i < res.totalPages; i++) {
      if (i == res.page) {
        pages.push(`<span class="font-black">${i}</span>`)
      } else {
        pages.push(`<a href="/?search=${searchData}&page=${i}">${i}</a>`)
      }
    }
    pagination.innerHTML = pages;

  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
if (window.location.search != "" && window.location.search.includes("search")) {
  const searchParams = new URLSearchParams(window.location.search);
  Controller.searchQuery(searchParams.get("search"), searchParams.get("page"))
}

