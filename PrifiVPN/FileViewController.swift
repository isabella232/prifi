//
//  FileViewController.swift
//  PrifiVPN
//
//  Created on 7/12/18.
//

import UIKit

class FileViewController: UIViewController {
    @IBOutlet weak var segmentedControl: UISegmentedControl!
    @IBOutlet weak var textView: UITextView!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        segment(0)
    }
    
    func loadFile(_ url: URL?) {
        guard let url = url else {
            return
        }
        do {
            textView.text = try String(contentsOf: url)
        } catch {
            NSLog("Error reading file. \(error)")
        }
    }
    
    func segment(_ index: Int) {
        switch index {
        case 0:
            loadFile(Bundle.main.url(forResource: "prifi", withExtension: ".toml"))
        case 1:
            loadFile(Bundle.main.url(forResource: "identity", withExtension: ".toml"))
        case 2:
            loadFile(Bundle.main.url(forResource: "group", withExtension: ".toml"))
        default:
            NSLog("Invalid segment")
        }
    }
    
    @IBAction func segmentSelected(_ sender: UISegmentedControl) {
        segment(sender.selectedSegmentIndex)
    }
}
