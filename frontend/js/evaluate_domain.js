var app2 = new Vue({
  el: '#app2',
  data: {
    evaluate: '',
    justCreated: true,
    domainEvaluation: {},
    visibleKeys: [],
    path: 'http://localhost:3000/domainEvaluations/'
  },
  methods: {
    evaluateDomain: function () {
      let domain = this.evaluate
      $.ajax({
        url: this.path + domain,
        type: 'GET',
        dataType: 'text',
        crossDomain: true,
        success: function(result) {
          app2.justCreated = false;
          app2.domainEvaluation = JSON.parse(result);
          app2.visibleKeys = Object.keys(app2.domainEvaluation.evaluation);
          app2.visibleKeys.splice(1,1)
        },
        error: function(error) {
        }
      });
    }
  }
})
