(function () {
    'use strict';

    angular
        .module('login')
        .controller('LoginController', LoginController);

    LoginController.$inject = ['$scope', '$window', 'dataservice', 'FlashService'];

    function LoginController($scope, $window, dataservice, FlashService) {

        activate();
        function activate() {
            console.log("Acitvated Login Controller");
            $scope.flash = FlashService;
        }

        $scope.login = function login() {

            dataservice.auth($scope.password).then(function (data) {
                if (data.status == 400) {
                    FlashService.setMessage(data.data);
                } else if (data.status == 200) {
                    //$window.location.href = 'http://10.0.0.4:4030/static/index.html';
                    $window.location.href = 'http://10.80.65.21:4030/static/index.html';
                }
            });

        }
    }
})();
