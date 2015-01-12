/* All angular application controllers */
var seagullServices = angular.module('seagullServices', []);

/* Docker service requests beego API server */
seagullServices.service('dockerService', ['$http', '$q',
    function ($http, $q) {
        var baseURL = '/dockerregistryapi/';
        var getVersionURI = baseURL + 'version';
        var getInfoURI = baseURL + 'info';
        var getImagesURI = baseURL + 'images/json';
        var getImageBaseURI = baseURL + 'images/';
        var searchImagesURI = baseURL + 'images/search';

        function getDataForURI(uri, params) {
            var deferred = $q.defer();
            var query = params ? '?' + params : "";
            var url = uri + query;
            $http.get(url).success(function (data) {
                deferred.resolve(data);
            }).error(function (reason) {
                deferred.reject(reason);
            });
            return deferred.promise;
        }

        function deleteDataForURI(uri, params){
            var deferred = $q.defer();
            var query = params ? '?' + params : "";
            var url = uri + query;
            $http.delete(url).success(function (data) {
                deferred.resolve(data);
            }).error(function (reason) {
                deferred.reject(reason);
            });
            return deferred.promise;
        }

        function getVersion() {
            return getDataForURI(getVersionURI);
        }

        function getInfo() {
            return getDataForURI(getInfoURI);
        }

        function getImages() {
            return getDataForURI(getImagesURI);
        }

        function getImageById(id) {
            return getDataForURI(getImageBaseURI + id + '/json');
        }

        function getImageByUserAndRepo(user, repo) {
            return getDataForURI(getImageBaseURI + user + "/" + repo + '/json');
        }

        function searchImages(term) {
            return getDataForURI(searchImagesURI, "term=" + term);
        }

        function deleteImage(image) {
            var name = image.Name;
            var tag = image.Tag;
            var uri = baseURL + 'repositories/' + name + '/tags/' + tag;
            return deleteDataForURI(uri);
            /*

            var deferred = $q.defer();
            $http({
                method: 'DELETE',
                url:
                data: '',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'}
            }).success(function () {
                if (status == 200) {
                    deferred.resolve(data);
                } else {
                    deferred.reject(status);
                }
            }).error(function (reason) {
                deferred.reject(reason);
            });
            return deferred.promise;
            */
        }

        return {
            getVersion: getVersion,
            getInfo: getInfo,
            getImages: getImages,
            getImageById: getImageById,
            getImageByUserAndRepo: getImageByUserAndRepo,
            deleteImage: deleteImage,
            searchImages: searchImages
        }

    }]);