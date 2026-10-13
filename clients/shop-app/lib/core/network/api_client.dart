import 'package:dio/dio.dart';
import 'package:quickeats_shop/core/constants/api_constants.dart';
import 'package:quickeats_shop/core/network/auth_interceptor.dart';
import 'package:quickeats_shop/services/auth_service.dart';

class ApiClient {
  late final Dio dio;

  ApiClient({required AuthService authService}) {
    dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {'Content-Type': 'application/json'},
    ));

    dio.interceptors.addAll([
      AuthInterceptor(authService: authService, dio: dio),
      LogInterceptor(requestBody: true, responseBody: true, logPrint: (o) => print('[API] $o')),
    ]);
  }
}
