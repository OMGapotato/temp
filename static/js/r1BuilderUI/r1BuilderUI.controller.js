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
        .filter('sysPackageFilter', sysPackageFilter)
        //.controller('RightCtrl', RightCtrl)

    function sysPackageFilter() {
        return function( items, condition, field) {
            var filtered = [];
            if (angular.isArray(condition)) {
                if (condition.length === 0) {
                    return items;
                }
            }
            if(condition === undefined || condition === '' || condition === null || condition.length === undefined){
                return items;
            }

            angular.forEach(items, function(item) {
                if (condition != null) {
                    switch (field) {
                        case 'Name':
                            if (condition === item.Name || item.Name ===null) {
                                filtered.push(item);
                            }
                            break;
                        case 'Job':
                            if (condition === item.Job || item.Job ===null) {
                                filtered.push(item);
                            }
                            break;
                        case 'Version':
                            angular.forEach(condition, function(version) {
                                if(version === item.Version || item.Version === null){
                                    filtered.push(item);
                                }
                            });
                            break;
                        default:
                    }
                }
            });

            return filtered;
        };
    }

    DialogController.$inject = ['$scope', 'dataservice','$mdDialog'];
    function DialogController($scope, dataservice, $mdDialog){

        activate();
        function activate() {
            $scope.progressVisible = "visible";
            $scope.selected = [];
            angular.forEach($scope.ctrl.locals.parent.selectedArtifacts, function(art) {
                $scope.selected.push(art);
            });
            getArtifacts();
            getSystemPackages();
        }

        $scope.setJobList = function setJobAndVersionList(sysPackage) {
            $scope.jobList = [];
            angular.forEach($scope.artifactList, function(artifact) {
                if ( ($scope.jobList.indexOf(artifact.Job) === -1) && (artifact.Name === sysPackage) ) {
                    $scope.jobList.push(artifact.Job);
                }
            });
        }

        $scope.setVersionList = function setVersionList(job, sysPackage) {
            $scope.versionList = [];
            angular.forEach($scope.artifactList, function(artifact) {
                if ( (job === artifact.Job) && ($scope.versionList.indexOf(artifact.Version) === -1) && (artifact.Name === sysPackage) ) {
                    $scope.versionList.push(artifact.Version); 
                }
            });
        }

        $scope.checkArtifact = function checkArtifact(artifact) {
            if (artifact.Job === "c247ufw-master") {
                return true;
            } else {
                return false;
            }
        }

        function getArtifacts() {
            dataservice.getArtifacts($scope.ctrl.locals.parent.selectedSystem).then(function(data) {
                $scope.artifactList = data.data;
                $scope.progressVisible = "hidden";
            });
        }

        function getSystemPackages() {
            dataservice.getSystemPackages($scope.ctrl.locals.parent.selectedSystem).then(function(data) {
                $scope.systemPackageList = data.data;
            });
        }

        $scope.toggle = function toggle(artifact,list) {
            var idx = -1;
            for(var i=0, len=list.length; i<len; i++) {
                if (list[i].Url === artifact.Url) {
                    idx = i;
                    break;
                }
            }
            
            if (idx > -1) {
                list.splice(idx, 1);
            } else {
                list.push(artifact);
            }
        };

        $scope.exists = function exists(artifact, list) {
            var idx = 0;
            for(var i=0, len=list.length; i<len; i++) {
                if (list[i].Url === artifact.Url) {
                    idx = i+1;
                    break;
                }
            }
            //console.log('exists = ' + idx);
            return idx;
        };

        $scope.addSelectedArtifacts = function addSelectedArtifacts(selectedArtifacts) {
            $mdDialog.hide();
            $scope.ctrl.locals.parent.addSelectedArtifacts(selectedArtifacts);
        }

        $scope.cancel = function cancel() {
            $mdDialog.cancel();
        }

    }

    R1BuilderUIController.$inject = ['$scope', '$q', '$timeout', '$interval', 'dataservice', '$mdSidenav', '$mdDialog', '$log', '$window'];
    function R1BuilderUIController($scope, $q, $timeout, $interval, dataservice, $mdSidenav, $mdDialog, $log, $window) {
       
        activate();
        function activate() {
            $scope.building = false;
            $scope.versionNumbers = ['0','1','2','3','4','5','6','7','8', '9'];
            $scope.buildNumbers = ['0','1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20',
                                    '21','22','23','24','25','26','27','28','29','30','31','32','33','34','35','36','37','38','39','40',
                                    '41','42','43','44','45','46','47','48','49','50','51','52','53','54','55','56','57','58','59','60',
                                    '61','62','63','64','65','66','67','68','69','70','71','72','73','74','75','76','77','78','79','80',
                                    '81','82','83','84','85','86','87','88','89','90','91','92','93','94','95','96','97','98','99'];
            getSystems();
            getBuildVersions();
        }

        $interval(function() {
            getBuildVersions();
            if ($scope.building) {
                angular.forEach($scope.buildVersions, function(version) {
                    if ($scope.versionBuilding === version) {
                        $scope.building = false;
                        $scope.versionBuilding = "";
                    }
                });
            }
        }, 2000);

        function getBuildVersions() {
            dataservice.getBuildVersions().then(function(data) {
                $scope.buildVersions = data.data;
            });
        }

        function getSystems() {
            dataservice.getSystems().then(function(data) {
                var systems = [];
                systems = data.data;
                
                $scope.systems = []

                angular.forEach(systems, function(system) {
                    var temp = {
                        Name: system.System,
                        SelectedArtifacts: []
                    };
                    $scope.systems.push(temp);
                });
                //console.log($scope.systems);
            });
        }

        $scope.getVersionPackages = function getVersionPackages(version) {
            dataservice.getVersionPackages(version).then(function(data) {
                console.log(data.data);
                $scope.systems = data.data.SysPackageList;
                var array = data.data.Version.split('-');
                $scope.buildNum = array[1];
                var tmpVersion = array[0].split('.');
                $scope.major = tmpVersion[0];
                $scope.minor = tmpVersion[1];
                $scope.patch = tmpVersion[2];
            });
        }

        $scope.downloadVersion = function downloadVersion(version) {
            $window.open('http://10.80.65.21:4030/rest/build/' + version + '/download');
            /*
            dataservice.downloadVersion(version).then(function(data) {
                
            });
            */
        }

        $scope.systemPackages = function systemPackages(system) {
            dataservice.getSystemPackages(system).then(function(data) {
                $scope[system + ".packages"] = data.data;
            });
        }
        $scope.getSystemNames = function getSystemNames(system) {
            return $scope[system + ".packages"];
        }

        $scope.exists = function exists(sysPackage, list) {
            var idx = 0;
            for(var i=0, len=list.length; i<len; i++) {
                if (list[i].Name === sysPackage) {
                    idx = i+1;
                    break;
                }
            }
            //console.log('exists = ' + idx);
            return idx;
        }

        $scope.addSelectedArtifacts = function addSelectedArtifacts(selectedArtifacts) {
            angular.forEach($scope.systems, function(system) {
                if (system.Name === $scope.selectedSystem) {
                    system.SelectedArtifacts = selectedArtifacts;
                } 
            });
        }

        $scope.build = function build(ev) {
            if ($scope.major == null || $scope.minor == null || $scope.patch == null || $scope.buildNum == null) {
                $mdDialog.show(
                  $mdDialog.alert()
                    .parent(angular.element(document.querySelector('#popupContainer')))
                    .clickOutsideToClose(true)
                    .title('Warning')
                    .textContent('You must specify a version.')
                    .ariaLabel('Alert Dialog')
                    .ok('Got it!')
                    .targetEvent(ev)
                );
            } else {

                var version = $scope.major + "." + $scope.minor + "." + $scope.patch + "-" + $scope.buildNum

                var found;
                angular.forEach($scope.buildVersions, function(previousVersion) {
                    if (version === previousVersion) {
                        found = true;
                        $mdDialog.show(
                          $mdDialog.alert()
                            .parent(angular.element(document.querySelector('#popupContainer')))
                            .clickOutsideToClose(true)
                            .title('Warning')
                            .textContent('That Version already exists. Please choose another one.')
                            .ariaLabel('Alert Dialog')
                            .ok('Got it!')
                            .targetEvent(ev)
                        );
                        return;
                    }
                });
                /*
                if ($scope.version = version) {
                }
                */
                if (!found) {
                    $scope.building = true;
                    dataservice.processBuild(version, $scope.systems).then(function(data) {
                        $scope.versionBuilding = version;
                        console.log(data);
                    });
                }
            }
        }

        $scope.deleteArtifact = function deleteArtifact(artifact, artifactList) {
            var index = artifactList.indexOf(artifact);
            artifactList.splice(index, 1);
        }

        $scope.showSelectArtifactPrompt = function showSelectArtifactPrompt(ev, selectedSystem, selectedArtifacts) {
            $scope.selectedSystem = selectedSystem;
            $scope.selectedArtifacts = selectedArtifacts;
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
        
        $scope.toggleLeft = buildDelayedToggler('left');
        
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
})();
