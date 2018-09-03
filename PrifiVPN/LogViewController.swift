//
//  LogViewController.swift
//  PrifiVPN
//
//  Created on 7/12/18.
//

import UIKit
import NetworkExtension

class LogViewController: UIViewController {
    @IBOutlet weak var logView: UITextView!
    @IBOutlet weak var pauseButton: UIButton!

    var appDelegate: AppDelegate!
    var prifiVpn: PrifiVpn!
    
    var paused = false
    var timer: Timer?
    var managers = [NETunnelProviderManager]()
    var currentRelay: NETunnelProviderManager?

    override func viewDidLoad() {
        super.viewDidLoad()
        NSLog("Loaded log view")
    }

    override func viewDidAppear(_ animated: Bool) {
        super.viewDidAppear(animated)
        NSLog("Log appeared")
        reloadManagers()
    }
    
    func reloadManagers() {
        stopLog()
        NETunnelProviderManager.loadAllFromPreferences { (managers, error) in
            if let error = error {
                NSLog("Error getting managers: \(error)")
                return
            }
            
            guard let managers = managers else {
                self.managers = [NETunnelProviderManager]()
                self.currentRelay = nil
                return
            }
            
            self.managers = managers
            self.currentRelay = managers.filter({ $0.isEnabled }).first
            self.startLog()
        }
    }

    private func startLog() {
        if paused {
            return
        }
        NSLog("startingLog")
        timer = Timer.scheduledTimer(timeInterval: 0.5, target: self,
                                     selector: #selector(logHandler),
                                     userInfo: nil, repeats: true)
    }

    private func stopLog() {
        NSLog("Stopping log")
        timer?.invalidate()
        timer = nil
    }

    override func viewWillDisappear(_ animated: Bool) {
        super.viewWillDisappear(animated)
        stopLog()
        NSLog("Log will disappear")
    }
    
    @objc func logHandler() {
        guard let manager = currentRelay, manager.connection.status == .connected else {
            return
        }

        NSLog("Getting log for \(manager.localizedDescription ?? "")")

        let session = manager.connection as! NETunnelProviderSession
        do {
            try session.sendProviderMessage(Data()) { response in
                if let response = response {
                    if let log = NSKeyedUnarchiver.unarchiveObject(with: response) as? [String] {
                        for line in log {
                            self.logView.text.append("\(line)\n")
                        }
                        let end = self.logView.text.count - 1
                        self.logView.scrollRangeToVisible(NSMakeRange(end, 1))
                    }
                }
            }
        } catch {
            NSLog("Could not establish communications channel with extension. Error: \(error)")
        }
    }


    @IBAction func clearLog(_ sender: UIButton) {
        logView.text = ""
    }


    @IBAction func pauseResumeLog(_ sender: UIButton) {
        paused = !paused
        if paused {
            pauseButton.setTitle("Resume", for: .normal)
            stopLog()
        } else {
            pauseButton.setTitle("Pause", for: .normal)
            startLog()
        }
    }
}
