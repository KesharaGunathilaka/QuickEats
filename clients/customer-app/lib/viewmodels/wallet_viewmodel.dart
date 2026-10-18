import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/wallet.dart';
import 'package:quickeats_customer/services/wallet_service.dart';

class WalletViewModel extends ChangeNotifier {
  final WalletService _walletService;
  WalletViewModel({required WalletService walletService}) : _walletService = walletService;

  Wallet? _wallet;
  bool _isLoading = false;
  String? _error;

  Wallet? get wallet => _wallet;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> loadWallet() async {
    _isLoading = true; notifyListeners();
    try { _wallet = await _walletService.getWallet(); } catch (e) { _error = 'Failed to load wallet'; }
    _isLoading = false; notifyListeners();
  }

  Future<bool> topUp(double amountLKR) async {
    try { await _walletService.topUp((amountLKR * 100).round()); await loadWallet(); return true;
    } catch (e) { _error = 'Top-up failed'; notifyListeners(); return false; }
  }
}
