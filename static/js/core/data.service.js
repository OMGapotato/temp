(function () {
    'use strict';

    angular
        .module('app.core')
        .factory('dataservice', dataservice)
        .factory('FlashService', FlashService)
        .factory('Base64', Base64);

    dataservice.$inject = ['$http', '$q', '$rootScope', '$cookieStore', 'Base64'];

    function dataservice($http, $q, $rootScope, $cookieStore, Base64, exception, logger) {

        var server = 'http://10.80.65.21:4030/rest';
        //var server = 'http://10.0.0.4:4030/rest';

        var service = {
            auth:               auth,
            getSystems:         getSystems,
            getSystemPackages:  getSystemPackages,
            getArtifacts:       getArtifacts,
            processBuild:       processBuild,
            getBuildVersions:   getBuildVersions,
            getVersionPackages: getVersionPackages,
            downloadVersion:    downloadVersion
        };

        return service;

        function auth(password) {
            var authdata = Base64.encode(password);

            $http.defaults.headers.common['Authorization'] = 'Basic ' + authdata;
            console.log(authdata);

            return $http({
                    method: 'GET',
                    url: server + '/auth'
                })
                .then(function(data, status, headers, config) {
                    $rootScope.globals = {
                        currentUser: {
                            authdata: authdata
                        }
                    };

                    $cookieStore.put('globals', $rootScope.globals);

                    return data;

                }, function(error) {
                    console.log('XHR failed for authentication');
                    console.log(error);
                    return error;
                });
        }

        function getSystems() {
            return $http({
                method: 'GET',
                url: server + '/systems'
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for getting cloud systems');
                console.log(error);
                return error;
            });
        }

        function getSystemPackages(system) {
            return $http({
                method: 'GET',
                url: server + '/systems/' + system
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for getting system packages');
                console.log(error);
                return error;
            });
        }

        function getArtifacts(system) {
            return $http({
                method: 'GET',
                url: server + '/' + system + '/artifacts'
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for getting artifacts');
                console.log(error);
                return error;
            });
        }

        function processBuild(version, systemPackageList) {
            angular.toJson(systemPackageList);
            var build = {
                'Version': version,
                'SysPackageList': systemPackageList
            };
            return $http({
                method: 'POST',
                url: server + '/build',
                data: build
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for posting build information');
                console.log(error);
                return error;
            });
        }

        function getBuildVersions() {
            return $http({
                method: 'GET',
                url: server + '/build/versions'
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for getting build versions');
                console.log(error);
                return error;
            });
        }

        function getVersionPackages(version) {
            return $http({
                method: 'GET',
                url: server + '/build/' + version
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for getting build version packages');
                console.log(error);
                return error;
            });
        }

        function downloadVersion(version) {
            return $http( {
                method: 'GET',
                url: server + '/build/' + version + '/download'
            })
            .then(function(data, status, headers, config) {
                return data;
            }, function(error) {
                console.log('XHR failed for downloading release');
                console.log(error);
                return error;
            });
        }
    }

    function Base64(){

        var keyStr = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';

        var service = {
            encode:     encode,
            decode:     decode
        }

        return service;

        function encode(input) {
            var output = "";
            var chr1, chr2, chr3 = "";
            var enc1, enc2, enc3, enc4 = "";
            var i = 0;

            do {
                chr1 = input.charCodeAt(i++);
                chr2 = input.charCodeAt(i++);
                chr3 = input.charCodeAt(i++);

                enc1 = chr1 >> 2;
                enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
                enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
                enc4 = chr3 & 63;

                if (isNaN(chr2)) {
                    enc3 = enc4 = 64;
                } else if (isNaN(chr3)) {
                    enc4 = 64;
                }

                output = output +
                    keyStr.charAt(enc1) +
                    keyStr.charAt(enc2) +
                    keyStr.charAt(enc3) +
                    keyStr.charAt(enc4);
                chr1 = chr2 = chr3 = "";
                enc1 = enc2 = enc3 = enc4 = "";
            } while (i < input.length);

            return output;
        }

        function decode(input) {
            var output = "";
            var chr1, chr2, chr3 = "";
            var enc1, enc2, enc3, enc4 = "";
            var i = 0;

            var base64test = /[^A-Za-z0-9\+\/\=]/g;
            if (base64test.exec(input)) {
                window.alert("There were invalid base64 characters in the input text.\n" +
                        "Valid base64 characters are A-Z, a-z, 0-9, '+', '/',and '='\n" +
                        "Expect errors in decoding.");
            }
            input = input.replace(/[^A-Za-z0-9\+\/\=]/g, "");

            do {
                enc1 = keyStr.indexOf(input.charAt(i++));
                enc2 = keyStr.indexOf(input.charAt(i++));
                enc3 = keyStr.indexOf(input.charAt(i++));
                enc4 = keyStr.indexOf(input.charAt(i++));

                chr1 = (enc1 << 2) | (enc2 >> 4);
                chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
                chr3 = ((enc3 & 3) << 6) | enc4;

                output = output + String.fromCharCode(chr1);

                if (enc3 != 64) {
                    output = output + String.fromCharCode(chr2);
                }
                if (enc4 != 64) {
                    output = output + String.fromCharCode(chr3);
                }

                chr1 = chr2 = chr3 = "";
                enc1 = enc2 = enc3 = enc4 = "";

            } while (i < input.length);

            return output;
        }
    }

    function FlashService() {

        var queue = [];
        var currentMessage = "";

        var service = {
            setMessage:         setMessage,
            getMessage:         getMessage
        };

        return service;

        function setMessage(message) {
            currentMessage=message;
        }

        function getMessage() {
            return currentMessage;
        }
    }

})();
