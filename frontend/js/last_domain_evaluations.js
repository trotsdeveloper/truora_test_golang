var app = new Vue({
  el: '#app',
  data: {
    domainEvaluations: {},
    path: 'http://localhost:3000/domainEvaluations/'
  },
  created: function () {
    this.loadTable()
  },
  methods: {
    loadTable: function () {
      $.ajax({
        url: this.path,
        type: 'GET',
        dataType: 'text',
        crossDomain: true,
        success: function(result) {
          app.domainEvaluations = JSON.parse(result)
        },
        error: function (error) {
        }
      });
    }
  }
})
