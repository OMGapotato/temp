(function() {
    'use strict';

    angular
        .module('app.r1BuilderUI')
        .config(function ($mdThemingProvider) {
                $mdThemingProvider.theme('lightTheme')
                    .primaryPalette('light-blue')
            })
        .controller('R1BuilderUIController', R1BuilderUIController)
        .controller('DialogController', DialogController)
        .controller('LeftCtrl', LeftCtrl)
        //.controller('RightCtrl', RightCtrl)

    DialogController.$inject = ['$scope'];
    function DialogController($scope){

        console.log($scope.ctrl.locals.parent.artifactList);
        $scope.artifactList = $scope.ctrl.locals.parent.artifactList;
        $scope.selectedArtifacts = [];
        //$scope.ctrl.locals.parent.dialog = "set this from dialog controller";
    }

    R1BuilderUIController.$inject = ['$scope', '$q', '$timeout', '$interval', 'dataservice', '$mdSidenav', '$mdDialog', '$log'];
    function R1BuilderUIController($scope, $q, $timeout, $interval, dataservice, $mdSidenav, $mdDialog, $log) {
       
        activate();
        function activate() {
            $scope.testing = "Testing";
            getSystems();
            getArtifacts();
        }
 
        function getSystems() {
            dataservice.getSystems().then(function(data) {
                $scope.jobs = data.data; 
            });
        }

        function getArtifacts() {
            dataservice.getArtifacts().then(function(data) {
                $scope.completeList = data.data;
                var len = $scope.completeList.length;
                $scope.artifactList = [];
                for (var i=0; i<len; i++) {
                    var buildLen = $scope.completeList[i].BuildHistory.length;
                    for (var j=0; j<buildLen; j++) {
                        var artLen = $scope.completeList[i].BuildHistory[j].artifacts.length;
                        for (var h=0; h<artLen; h++) {
                            $scope.artifactList.push($scope.completeList[i].BuildHistory[j].artifacts[h]);
                        }
                    }
                }//end for
            });
        }

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

        $scope.showArtifactList = function showArtifactList() {
            console.log($scope.dialog);
        }

        $scope.showPrompt = function showPrompt(ev) {

            $mdDialog.show({
                locals: {parent: $scope},
                controller: DialogController,
                controllerAs: 'ctrl',
                bindToController: true,
                templateUrl: 'dialog.tmpl.html',
                parent: angular.element(document.body),
                targetEvent: ev,
                clickOutsideToClose:true,
                fullscreen: $scope.customFullscreen // Only for -xs, -sm breakpoints.
            })
            .then(function(answer) {
                $scope.status = 'You said the information was "' + answer + '".';
            }, function() {
                $scope.status = 'You cancelled the dialog.';
            });
        }

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
