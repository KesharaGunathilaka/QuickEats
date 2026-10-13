import 'package:dio/dio.dart';
import 'package:quickeats_customer/services/auth_service.dart';

class AuthInterceptor extends Interceptor {
  final AuthService authService;
  final Dio dio;
  AuthInterceptor({required this.authService, required this.dio});

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = authService.accessToken;
    if (token != null) options.headers['Authorization'] = 'Bearer $token';
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401 && authService.refreshToken != null) {
      try {
        if (await authService.refreshTokens()) {
          final opts = err.requestOptions;
          opts.headers['Authorization'] = 'Bearer ${authService.accessToken}';
          handler.resolve(await dio.fetch(opts));
          return;
        }
      } catch (_) { authService.logout(); }
    }
    handler.next(err);
  }
}
