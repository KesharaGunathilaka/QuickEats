import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/models/wallet.dart';

class WalletService {
  final ApiClient _api;
  WalletService({required ApiClient apiClient}) : _api = apiClient;

  Future<Wallet> getWallet() async {
    final resp = await _api.dio.get(ApiConstants.wallet);
    return Wallet.fromJson(resp.data['data']);
  }

  Future<void> topUp(int amountCents) async {
    await _api.dio.post(ApiConstants.walletTopup, data: {'amount': amountCents});
  }
}
