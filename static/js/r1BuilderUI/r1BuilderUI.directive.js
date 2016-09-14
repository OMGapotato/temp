(function() {
    'use strict';

    angular
        .module('app.r1BuilderUI')
        .directive('job', job)

    function job() {
        return {
            retrict: 'E',
            scope: false,
            templateUrl: 'job.tmpl.html'
        }
    }

})();
