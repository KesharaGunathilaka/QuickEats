import 'package:dio/dio.dart';
import 'package:quickeats_shop/services/auth_service.dart';

class AuthInterceptor extends Interceptor {
  final AuthService authService;
  final Dio dio;

  AuthInterceptor({required this.authService, required this.dio});

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = authService.accessToken;
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401 && authService.refreshToken != null) {
      try {
        final refreshed = await authService.refreshTokens();
        if (refreshed) {
          // Retry the failed request with new token
          final options = err.requestOptions;
          options.headers['Authorization'] = 'Bearer ${authService.accessToken}';
          final response = await dio.fetch(options);
          handler.resolve(response);
          return;
        }
      } catch (_) {
        authService.logout();
      }
    }
    handler.next(err);
  }
}
