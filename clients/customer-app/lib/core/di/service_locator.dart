import 'package:get_it/get_it.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/core/network/ws_client.dart';
import 'package:quickeats_customer/services/account_service.dart';
import 'package:quickeats_customer/services/order_notification_service.dart';
import 'package:quickeats_customer/viewmodels/settings_viewmodel.dart';
import 'package:quickeats_customer/services/auth_service.dart';
import 'package:quickeats_customer/services/shop_service.dart';
import 'package:quickeats_customer/services/order_service.dart';
import 'package:quickeats_customer/services/card_service.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/shops_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/card_viewmodel.dart';

final getIt = GetIt.instance;

Future<void> setupServiceLocator() async {
  final authService = AuthService();
  await authService.init();
  getIt.registerSingleton<AuthService>(authService);
  getIt.registerSingleton<ApiClient>(ApiClient(authService: authService));

  getIt.registerSingleton<ShopService>(ShopService(apiClient: getIt<ApiClient>()));
  getIt.registerSingleton<OrderService>(OrderService(apiClient: getIt<ApiClient>()));
  getIt.registerSingleton<CardService>(CardService(apiClient: getIt<ApiClient>()));
  getIt.registerSingleton<AccountService>(AccountService(apiClient: getIt<ApiClient>()));

  // Live order updates: connects on login, disconnects on logout
  final notifications = OrderNotificationService(authService: authService, wsClient: WSClient());
  await notifications.init();
  getIt.registerSingleton<OrderNotificationService>(notifications);

  getIt.registerFactory<AuthViewModel>(() => AuthViewModel(authService: getIt<AuthService>()));
  getIt.registerFactory<ShopsViewModel>(() => ShopsViewModel(shopService: getIt<ShopService>()));
  getIt.registerSingleton<CartViewModel>(CartViewModel()); // singleton — persists across screens
  getIt.registerFactory<OrderViewModel>(() => OrderViewModel(
        orderService: getIt<OrderService>(), notifications: getIt<OrderNotificationService>()));
  getIt.registerFactory<CardViewModel>(() => CardViewModel(cardService: getIt<CardService>()));
  getIt.registerFactory<SettingsViewModel>(() => SettingsViewModel(
        accountService: getIt<AccountService>(), notifications: getIt<OrderNotificationService>()));
}
