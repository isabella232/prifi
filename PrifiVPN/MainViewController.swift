//
//  MainView.swift
//  PrifiVPN
//
//  Created on 8/10/18.
//  Copyright © 2018 dedis. All rights reserved.
//

import UIKit
import CocoaAsyncSocket
import Reachability
import CoreActionSheetPicker
import NetworkExtension

class MainViewController: UIViewController, GCDAsyncSocketDelegate {
    // MARK: Outlets
    @IBOutlet weak var powerButton: UIButton!
    @IBOutlet weak var ssidLabel: UILabel!
    @IBOutlet weak var relayButton: UIButton!
    @IBOutlet weak var relayAddress: UILabel!
    @IBOutlet weak var lockImage: UIImageView!
    
    @IBOutlet weak var buttonOverlay: UIView!
    @IBOutlet weak var statusLabel: UILabel!
    @IBOutlet weak var bytesLabel: UILabel!
    
    var prifiVpn: PrifiVpn!
    var appDelegate: AppDelegate!
    
    var managers = [NETunnelProviderManager]()
    
    var configurationRepository: ConfigurationRepository!
    var groupRepository: GroupRepository!
    
    var currentGroup: Group? {
        didSet {
            ssidLabel.text = currentGroup?.name ?? "None"
        }
    }

    var currentRelay: NETunnelProviderManager? {
        didSet {
            var relayText = "None"
            if let group = currentGroup?.name, let relay = currentRelay?.localizedDescription {
                relayText = "\(group) @ \(relay)"
            }
            relayButton.setTitle(relayText, for: .normal)
            powerButton.isEnabled = currentRelay != nil
            if let relayAddress = currentRelay?.protocolConfiguration?.serverAddress {
                self.relayAddress.text = relayAddress
            } else {
                self.relayAddress.text = ""
            }
            
            currentRelay?.isEnabled = true
            currentRelay?.saveToPreferences { error in
                if let error = error {
                    NSLog("Failed to save configuration: \(error)")
                    return
                }
                if let configName = self.currentRelay?.localizedDescription {
                    NSLog("Enabled: \(configName)")
                }
            }
        }
    }
    
    var relayTest: NetworkTools.Request? = nil
    
    var reachability: Reachability!
    
    var isServiceRunning: Bool = false {
        didSet {
            powerButton.isRunning(isServiceRunning)
            showOverlay(isServiceRunning)
            lockImage.adjustsImageSizeForAccessibilityContentSizeCategory = true
            lockImage.image = isServiceRunning ? #imageLiteral(resourceName: "ic_lock_on") : #imageLiteral(resourceName: "ic_lock_off")
            buttonOverlay.isHidden = !isServiceRunning
            statusLabel.isHidden = !isServiceRunning
            bytesLabel.isHidden = !isServiceRunning
        }
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        self.appDelegate = UIApplication.shared.delegate as! AppDelegate
        self.configurationRepository = ConfigurationRepository()
    }
    
    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        powerButton.layer.cornerRadius = powerButton.frame.width * 0.5
        powerButton.isRunning(isServiceRunning)
    }
    
    func showOverlay(_ isServiceRunning: Bool) {
    }
    
    func reloadManagers() {
        self.groupRepository = GroupRepository()
        self.currentGroup = groupRepository.activeGroup
        NETunnelProviderManager.loadAllFromPreferences { (managers, error) in
            if let error = error {
                NSLog("Error getting managers: \(error)")
                return
            }
            
            guard let currentGroup = self.currentGroup, let managers = managers else {
                self.managers = [NETunnelProviderManager]()
                self.currentRelay = nil
                return
            }
            
            self.managers = managers.filter{ $0.groupId == currentGroup.id }
            self.currentRelay = self.managers.filter({ $0.isEnabled }).first
        }
    }
    
    func startNotifier() {
        NotificationCenter.default.addObserver(self,
                                               selector: #selector(reachabilityChanged),
                                               name: Notification.Name.reachabilityChanged,
                                               object: nil)
        NotificationCenter.default.addObserver(self,
                                               selector: #selector(self.NEVPNStatusDidChange(_:)),
                                               name: NSNotification.Name.NEVPNStatusDidChange,
                                               object: nil)
        
        reachability = Reachability()!
        do {
            try self.reachability.startNotifier()
        } catch {
            NSLog("Error starting reachability: \(error.localizedDescription)")
        }
    }
    
    @objc func reachabilityChanged(_: Notification?) {
        let ssid = NetworkTools.getSsid() ?? "No Network"
//        ssidLabel.text = "\(ssid)"
        NSLog("SSID: \(ssid)")
    }
    

    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        reloadManagers()
        startNotifier()
        // Force initial update
        reachabilityChanged(nil)
        NEVPNStatusDidChange(nil)
    }
    
    override func viewWillLayoutSubviews() {
        powerButton.isRunning(false)
        relayButton.layer.cornerRadius = relayButton.frame.height / 2
        relayButton.clipsToBounds = true
    }
    
    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
        NotificationCenter.default.removeObserver(self)
    }
    
    func setupRequests() -> NetworkTools.Request? {
        guard let currentRelay = currentRelay else { return nil }
        let config = Configuration(from: currentRelay)
        
        guard let host = config.host, let relayPort = config.relayPort, let socksPort = config.socksPort else {
            return nil
        }
        
        return NetworkTools.Request(host: host, onPort: relayPort) { (sock, host, port, error) in
            if let error = error {
                NSLog(error.localizedDescription)
                let alert = UIAlertController(title: "Relay Unreachable", message: "Could not connect to relay", preferredStyle: .alert)
                alert.addAction(UIAlertAction(title: "OK", style: .default))
                self.present(alert, animated: true, completion: nil)
                return
            }
            NSLog("Relay working")
        }.then(NetworkTools.Request(host: host, onPort: socksPort) { (sock, host, port, error) in
            if let error = error {
                NSLog(error.localizedDescription)
                let alert = UIAlertController(title: "Socks Unreachable", message: "Could not connect to socks server", preferredStyle: .alert)
                alert.addAction(UIAlertAction(title: "OK", style: .default))
                self.present(alert, animated: true, completion: nil)
                return
            }
            NSLog("Relay and socks working!")
            self.startPrifiVpn()
        })
    }
    
    @objc func NEVPNStatusDidChange(_ notification: Notification?) {
        guard let manager = currentRelay else {
            return
        }
        let status = manager.connection.status
        
        NSLog("Status: \(status)")
        
        switch status {
        case .disconnecting, .connecting:
            powerButton.isUserInteractionEnabled = false
        default:
            powerButton.isUserInteractionEnabled = true
        }
        
        switch status {
        case .disconnected, .invalid:
            self.isServiceRunning = false
        default:
            self.isServiceRunning = true
        }
    }
    
    func startVpn() {
        relayTest = setupRequests()
        
        guard relayTest != nil else {
            let alert = UIAlertController(title: "Invalid configuration", message: "Pleases check your configurations.", preferredStyle: .alert)
            alert.addAction(UIAlertAction(title: "OK", style: .default))
            present(alert, animated: true)
            NSLog("Relay test is nil")
            return
        }
        relayTest!.make()
    }
    
    func stopVpn() {
        NSLog("Stopping VPN...")
        currentRelay?.connection.stopVPNTunnel()
    }
    
    // Will be called after the connectivity tests
    func startPrifiVpn() {
        guard let manager = currentRelay else {
            NSLog("No manager selected")
            return
        }
        NSLog("Starting Prifi VPN (\(manager.localizedDescription ?? ""))")
        
        // Update with autodisconnect current value
        let config = manager.toConfiguration()
        
        config.autoDisconnect = !GlobalSettings.prifiOnly
        config.save(to: manager) { (error) in
            if let error = error {
                fatalError("Could not update manager: \(error)")
            }
            do {
                try manager.connection.startVPNTunnel()
            } catch {
                NSLog("Could not start PriFi VPN. \(error)")
            }
        }
    }
    
    @IBAction func buttonTouch(_ sender: UIButton) {
        if (!isServiceRunning) {
            startVpn()
        } else {
            stopVpn()
        }
    }
    
    @IBAction func changeRelay(_ sender: UIButton) {
        guard managers.count > 0 else {
            let alert = UIAlertController(title: "No saved configurations", message: "Pleases create a configuration first.", preferredStyle: .alert)
            alert.addAction(UIAlertAction(title: "OK", style: .default))
            present(alert, animated: true)
            return
        }
        
        let currentRelayName = currentRelay?.localizedDescription ?? ""
        let configNames = managers.map({$0.localizedDescription ?? ""})
        let initial = configNames.index(of: currentRelayName) ?? 0
        
        let picker = ActionSheetMultipleStringPicker(
            title: "Choose Server",
            rows: [configNames],
            initialSelection: [initial],
            doneBlock: { _, index, value in
                if let i = index?[0] as? Int {
                   self.currentRelay = self.managers[i]
                }
                return
            },
            cancel: { ActionMultipleStringCancelBlock in
                return
            },
            origin: sender)
        
        picker?.pickerTextAttributes[NSAttributedStringKey.font] = UIFont.systemFont(ofSize: 19.0)
        picker?.show()
    }
}

fileprivate extension UIButton {
    func isRunning(_ isRunning: Bool) {
        layer.shadowOffset = CGSize(width: 0, height: 0)
        layer.shadowOpacity = 0.7
        layer.shadowColor = UIColor.darkGray.cgColor
        layer.borderColor = UIColor.darkGray.cgColor
        
        if isRunning {
            layer.shadowRadius = 5.0
            layer.borderWidth = 0
            backgroundColor = UIColor(named: "colorOn")
        } else {
            layer.shadowRadius = 20.0
            layer.borderWidth = 0.2
            backgroundColor = UIColor(named: "colorOff")
        }
    }
}
