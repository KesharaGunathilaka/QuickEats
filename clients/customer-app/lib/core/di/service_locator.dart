import 'package:get_it/get_it.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/services/auth_service.dart';
import 'package:quickeats_customer/services/shop_service.dart';
import 'package:quickeats_customer/services/order_service.dart';
import 'package:quickeats_customer/services/wallet_service.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/shops_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/wallet_viewmodel.dart';

final getIt = GetIt.instance;

Future<void> setupServiceLocator() async {
  final authService = AuthService();
  await authService.init();
  getIt.registerSingleton<AuthService>(authService);
  getIt.registerSingleton<ApiClient>(ApiClient(authService: authService));

  getIt.registerSingleton<ShopService>(ShopService(apiClient: getIt<ApiClient>()));
  getIt.registerSingleton<OrderService>(OrderService(apiClient: getIt<ApiClient>()));
  getIt.registerSingleton<WalletService>(WalletService(apiClient: getIt<ApiClient>()));

  getIt.registerFactory<AuthViewModel>(() => AuthViewModel(authService: getIt<AuthService>()));
  getIt.registerFactory<ShopsViewModel>(() => ShopsViewModel(shopService: getIt<ShopService>()));
  getIt.registerSingleton<CartViewModel>(CartViewModel()); // singleton — persists across screens
  getIt.registerFactory<OrderViewModel>(() => OrderViewModel(orderService: getIt<OrderService>()));
  getIt.registerFactory<WalletViewModel>(() => WalletViewModel(walletService: getIt<WalletService>()));
}
