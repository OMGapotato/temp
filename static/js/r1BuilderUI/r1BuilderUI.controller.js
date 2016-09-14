(function() {
    'use strict';

    angular
        .module('app.r1BuilderUI')
        .config(function ($mdThemingProvider) {
                $mdThemingProvider.theme('lightTheme')
                    .primaryPalette('light-blue')
            })
        .controller('R1BuilderUIController', R1BuilderUIController)
        .controller('LeftCtrl', LeftCtrl)
        //.controller('RightCtrl', RightCtrl)


    R1BuilderUIController.$inject = ['$scope', '$q', '$timeout', '$interval', 'dataservice', '$mdSidenav', '$log'];
    function R1BuilderUIController($scope, $q, $timeout, $interval, dataservice, $mdSidenav, $log) {
        
        $scope.outputBuild = function outputBuild(artifacts) {
            $scope.artifacts = artifacts;
        }

        $scope.deleteArtifact = function deleteArtifact(artifact, artifactList) {
            var index = artifactList.indexOf(artifact);
            artifactList.splice(index, 1);
        }

        $scope.deleteJob = function deleteJob(jobName, jobList) {
            var index = jobList.indexOf(jobName);
            jobList.splice(index, 1);
        }

        $scope.addArtifact = function addArtifact(name, artifactList) {
            artifactList.push(name + ' new artifact1');
        }

/*
        type Job struct {
            name        string
            build       int
            artifacts   []string
        }

        type Build struct {
            version    string
            tags       []string
            artifacts  []string
        }
*/
        // /rest/:job -> returns an array of job objects. object contains job name, build number, and list of artifacts
        //
        //    type Job struct {
        //        name        string
        //        build       int
        //        artifacts   []string
        //    }
        //
        // /rest/:job/:build -> returns list of artifacts for that job name and build number
        $scope.jobs = [
            {
                name: 'ServerBackup-5.14.0',
                build: '410',
                artifacts: ['idera-hotcopy-amd64-5.14.4.deb','r1soft-cdp-agent-amd64-5.14.4-434.deb','r1soft-cdp-async-agent-amd64-5.14.4-434.deb']
            },
            {
                name: 'ServerBackup-5.14.0',
                build: '420',
                artifacts: ['idera-hotcopy-amd64-5.14.6.deb','r1soft-cdp-agent-amd64-5.14.6-455.deb','r1soft-cdp-async-agent-amd64-5.14.6-455.deb']
            },
            {
                name: 'ServerBackup-5.14.0',
                build: '430',
                artifacts: ['idera-hotcopy-amd64-5.16.0.deb','r1soft-cdp-agent-amd64-5.16.0-466.deb','r1soft-cdp-async-agent-amd64-5.16.0-466.deb']
            },
            {
                name: 'BBS-2.0',
                build: '40',
                artifacts: ['r1rm-2.1.0-23.deb','ServerBackup-6.0.1-84.deb','r1ctl-2.1.0-23']
            },
            {
                name: 'BBS-2.0',
                build: '38',
                artifacts: ['r1rm-2.0.0-21.deb','ServerBackup-6.0.0-42.deb','r1ctl-2.0.0-21']
            },
        ];

        //  /rest/:job/builds -> returns all build versions
        //  /rest/:job/build/:version -> returns a list tags and artifacts of build version
        $scope.BBSBuilds = [
            {
                jobName: 'BBS-2.0',
                version: '0.1.1',
                tags: ['something', 'important'],
                artifacts: ['r1rm-2.1.0-23.deb','ServerBackup-6.0.0-42.deb','r1ctl-2.0.0-21']
            },
            {
                jobName: 'BBS-2.0',
                version: '0.1.2',
                tags: ['json', 'sometag'],
                artifacts: ['r1rm-2.1.0-23.deb','ServerBackup-6.0.0-42.deb','r1ctl-2.0.0-21']
            },
            {
                jobName: 'BBS-2.0',
                version: '0.1.3',
                tags: ['othertag', 'not'],
                artifacts: ['r1rm-2.1.0-23.deb','ServerBackup-6.0.0-42.deb','r1ctl-2.0.0-21']
            }
        ];

        $scope.SBMBuilds = [
            {
                jobName: 'ServerBackup-5.14.0',
                version: '0.2.1',
                tags: ['something', 'important'],
                artifacts: ['idera-hotcopy-amd64-5.16.0.deb','r1soft-cdp-agent-amd64-5.16.0-466.deb','r1soft-cdp-async-agent-amd64-5.16.0-466.deb']
            },
            {
                jobName: 'ServerBackup-5.14.0',
                version: '0.2.2',
                tags: ['json', 'sometag'],
                artifacts: ['idera-hotcopy-amd64-5.16.0.deb','r1soft-cdp-agent-amd64-5.16.0-466.deb','r1soft-cdp-async-agent-amd64-5.16.0-466.deb']
            },
            {
                jobName: 'ServerBackup-5.14.0',
                version: '0.2.3',
                tags: ['othertag', 'not'],
                artifacts: ['idera-hotcopy-amd64-5.16.0.deb','r1soft-cdp-agent-amd64-5.16.0-466.deb','r1soft-cdp-async-agent-amd64-5.16.0-466.deb']
            }

        ];



        $scope.toggleLeft = buildDelayedToggler('left');
        /*
        $scope.toggleRight = buildToggler('right');
        $scope.isOpenRight = function(){
            return $mdSidenav('right').isOpen();
        };
        */
        /**
         *      * Supplies a function that will continue to operate until the
         *           * time is up.
         *                */
        function debounce(func, wait, context) {
            var timer;

            return function debounced() {
                var context = $scope,
                args = Array.prototype.slice.call(arguments);
                $timeout.cancel(timer);
                timer = $timeout(function() {
                    timer = undefined;
                    func.apply(context, args);
                }, wait || 10);
            };
        }

        /**
         *      * Build handler to open/close a SideNav; when animation finishes
         *           * report completion in console
         *                */
        function buildDelayedToggler(navID) {
            return debounce(function() {
                // Component lookup should always be available since we are not using `ng-if`
                $mdSidenav(navID)
                    .toggle()
                    .then(function () {
                        $log.debug("toggle " + navID + " is done");
                    });
            }, 200);
        }

        function buildToggler(navID) {
            return function() {
                // Component lookup should always be available since we are not using `ng-if`
                $mdSidenav(navID)
                    .toggle()
                    .then(function () {
                        $log.debug("toggle " + navID + " is done");
                    });
            }
        }    

    }

    LeftCtrl.$inject = ['$scope', '$timeout', '$mdSidenav', '$log'];
    function LeftCtrl($scope, $timeout, $mdSidenav, $log) {
        $scope.close = function () {
            // Component lookup should always be available since we are not using `ng-if`
            $mdSidenav('left').close()
                .then(function () {
                    $log.debug("close LEFT is done");
                });

        };
    }
/*
    RightCtrl.$inject = ['$scope', '$timeout', '$mdSidenav', '$log'];
    function RightCtrl($scope, $timeout, $mdSidenav, $log) {
        $scope.close = function () {
            // Component lookup should always be available since we are not using `ng-if`
            $mdSidenav('right').close()
                .then(function () {
                    $log.debug("close RIGHT is done");
                });

        };
    }
*/
})();
